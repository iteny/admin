package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/str"
	"cms/internal/repository"
)

// ArticleInput 是文章创建/编辑表单的业务输入。
type ArticleInput struct {
	Title          string
	Slug           string
	Summary        string
	Content        string
	SubContent     string
	ContentType    string
	CoverURL       string
	CategoryID     uint64
	Status         string
	IsTop          bool
	IsRecommend    bool
	PublishedAt    *time.Time
	TagIDs         []uint64
	SeoTitle       string
	SeoKeywords    string
	SeoDescription string
}

// ArticleService 负责文章业务规则。
type ArticleService struct {
	articles   *repository.ArticleRepo
	categories *repository.CategoryRepo
	tags       *repository.TagRepo
}

// NewArticleService 创建文章服务。
func NewArticleService(articles *repository.ArticleRepo, categories *repository.CategoryRepo, tags *repository.TagRepo) *ArticleService {
	return &ArticleService{articles: articles, categories: categories, tags: tags}
}

// List 返回文章列表。
func (s *ArticleService) List(ctx context.Context, f repository.ArticleFilter) ([]model.Article, int64, error) {
	return s.articles.List(ctx, f)
}

// Get 按 ID 返回文章。
func (s *ArticleService) Get(ctx context.Context, id uint64) (*model.Article, error) {
	article, err := s.articles.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return article, nil
}

// GetPublished 按 slug 返回已发布文章。
func (s *ArticleService) GetPublished(ctx context.Context, slug string) (*model.Article, error) {
	article, err := s.articles.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return article, nil
}

// Create 创建文章。
func (s *ArticleService) Create(ctx context.Context, in ArticleInput, authorID uint64) (*model.Article, error) {
	article, err := s.build(ctx, in, 0, authorID)
	if err != nil {
		return nil, err
	}
	if err := s.articles.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	if err := s.replaceTags(ctx, article, in.TagIDs); err != nil {
		return nil, err
	}
	return s.articles.GetByID(ctx, article.ID)
}

// Update 更新文章。
func (s *ArticleService) Update(ctx context.Context, id uint64, in ArticleInput, authorID uint64) (*model.Article, error) {
	existing, err := s.articles.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	article, err := s.build(ctx, in, id, authorID)
	if err != nil {
		return nil, err
	}
	article.ID = existing.ID
	article.CreatedAt = existing.CreatedAt
	article.AuthorID = existing.AuthorID
	if err := s.articles.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}
	if err := s.replaceTags(ctx, article, in.TagIDs); err != nil {
		return nil, err
	}
	return s.articles.GetByID(ctx, article.ID)
}

// Delete 删除文章。
func (s *ArticleService) Delete(ctx context.Context, id uint64) error {
	if err := s.articles.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return s.tags.RecountAll(ctx)
}

// Batch 批量操作文章：delete / publish / archive。
func (s *ArticleService) Batch(ctx context.Context, ids []uint64, action string) error {
	if len(ids) == 0 {
		return apperr.NewValidation("ids", "请选择文章")
	}
	switch action {
	case "delete":
		if err := s.articles.DeleteMany(ctx, ids); err != nil {
			return err
		}
		return s.tags.RecountAll(ctx)
	case "publish", "archive":
		status := model.ArticleStatusPublished
		if action == "archive" {
			status = model.ArticleStatusArchived
		}
		fields := map[string]any{"status": status}
		if status == model.ArticleStatusPublished {
			fields["published_at"] = time.Now().UTC()
		}
		for _, id := range ids {
			if err := s.articles.UpdateFields(ctx, id, fields); err != nil {
				return err
			}
		}
		return nil
	default:
		return apperr.NewValidation("action", "未知的批量操作")
	}
}

// View 增加浏览量。
func (s *ArticleService) View(ctx context.Context, id uint64) error {
	return s.articles.IncrementView(ctx, id)
}

// Related 返回相关文章。
func (s *ArticleService) Related(ctx context.Context, categoryID, excludeID uint64) ([]model.Article, error) {
	return s.articles.Related(ctx, categoryID, excludeID, 4)
}

// PrevNext 返回上一篇/下一篇，不存在时对应值为 nil。
func (s *ArticleService) PrevNext(ctx context.Context, article *model.Article) (*model.Article, *model.Article, error) {
	prev, next, err := s.articles.PrevNext(ctx, article)
	if err != nil {
		return nil, nil, err
	}
	if prev != nil && prev.ID == 0 {
		prev = nil
	}
	if next != nil && next.ID == 0 {
		next = nil
	}
	return prev, next, nil
}

func (s *ArticleService) build(ctx context.Context, in ArticleInput, excludeID, authorID uint64) (*model.Article, error) {
	title := SanitizeText(in.Title)
	if title == "" {
		return nil, apperr.NewValidation("title", "标题不能为空")
	}
	if len([]rune(title)) > 200 {
		return nil, apperr.NewValidation("title", "标题不能超过 200 个字符")
	}
	if in.CategoryID == 0 {
		return nil, apperr.NewValidation("category_id", "请选择分类")
	}
	if _, err := s.categories.GetByID(ctx, in.CategoryID); err != nil {
		return nil, apperr.NewValidation("category_id", "分类不存在")
	}

	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = str.Slug(title, true)
	} else {
		slug = str.NormalizeSlug(slug)
	}
	slug, err := s.uniqueSlug(ctx, slug, excludeID)
	if err != nil {
		return nil, err
	}

	status := model.ArticleStatus(in.Status)
	switch status {
	case model.ArticleStatusPublished, model.ArticleStatusArchived:
	case model.ArticleStatusDraft:
		// 历史草稿或旧页面提交的 draft 转为归档，避免意外发布。
		status = model.ArticleStatusArchived
	default:
		status = model.ArticleStatusPublished
	}
	contentType := model.ContentType(in.ContentType)
	if contentType != model.ContentTypeMarkdown {
		contentType = model.ContentTypeHTML
	}
	publishedAt := in.PublishedAt
	if status == model.ArticleStatusPublished && publishedAt == nil {
		now := time.Now().UTC()
		publishedAt = &now
	}
	summary := SanitizeText(in.Summary)
	if summary == "" {
		summary = str.Truncate(stripTags(SanitizeHTML(in.Content)), 160)
	}

	article := &model.Article{
		Title:          title,
		Slug:           slug,
		Summary:        summary,
		Content:        SanitizeHTML(in.Content),
		SubContent:     SanitizeHTML(in.SubContent),
		ContentType:    contentType,
		CoverURL:       strings.TrimSpace(in.CoverURL),
		CategoryID:     in.CategoryID,
		AuthorID:       authorID,
		Status:         status,
		IsTop:          in.IsTop,
		IsRecommend:    in.IsRecommend,
		PublishedAt:    publishedAt,
		SeoTitle:       SanitizeText(in.SeoTitle),
		SeoKeywords:    SanitizeText(in.SeoKeywords),
		SeoDescription: SanitizeText(in.SeoDescription),
	}
	return article, nil
}

func (s *ArticleService) replaceTags(ctx context.Context, article *model.Article, tagIDs []uint64) error {
	tags, err := s.tags.FindByIDs(ctx, tagIDs)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	if err := s.articles.SetTags(ctx, article, tags); err != nil {
		return fmt.Errorf("set tags: %w", err)
	}
	return s.tags.RecountAll(ctx)
}

func (s *ArticleService) uniqueSlug(ctx context.Context, base string, excludeID uint64) (string, error) {
	if base == "" {
		base = "article"
	}
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		exists, err := s.articles.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("无法生成唯一 slug")
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
