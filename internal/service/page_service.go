package service

import (
	"context"
	"fmt"
	"strings"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
	"cms/internal/pkg/str"
	"cms/internal/repository"
)

// PageInput 是单页表单输入。
type PageInput struct {
	Title          string
	Slug           string
	Content        string
	Status         string
	Sort           int
	SeoTitle       string
	SeoKeywords    string
	SeoDescription string
}

// PageService 负责单页业务。
type PageService struct {
	repo *repository.PageRepo
}

// NewPageService 创建单页服务。
func NewPageService(repo *repository.PageRepo) *PageService { return &PageService{repo: repo} }

// List 返回单页列表。
func (s *PageService) List(ctx context.Context, q pagination.Query) ([]model.Page, int64, error) {
	return s.repo.List(ctx, q)
}

// Get 按 ID 获取单页。
func (s *PageService) Get(ctx context.Context, id uint64) (*model.Page, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// GetBySlug 按 slug 获取已发布单页。
func (s *PageService) GetBySlug(ctx context.Context, slug string) (*model.Page, error) {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// Published 返回前台可展示的单页。
func (s *PageService) Published(ctx context.Context) ([]model.Page, error) {
	return s.repo.Published(ctx)
}

// Create 创建单页。
func (s *PageService) Create(ctx context.Context, in PageInput) (*model.Page, error) {
	item, err := s.build(ctx, in, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	return item, nil
}

// Update 更新单页。
func (s *PageService) Update(ctx context.Context, id uint64, in PageInput) (*model.Page, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	item, err := s.build(ctx, in, id)
	if err != nil {
		return nil, err
	}
	item.ID = id
	item.CreatedAt = existing.CreatedAt
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update page: %w", err)
	}
	return item, nil
}

// Delete 删除单页。
func (s *PageService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete page: %w", err)
	}
	return nil
}

func (s *PageService) build(ctx context.Context, in PageInput, excludeID uint64) (*model.Page, error) {
	title := SanitizeText(in.Title)
	if title == "" {
		return nil, apperr.NewValidation("title", "标题不能为空")
	}
	if len([]rune(title)) > 100 {
		return nil, apperr.NewValidation("title", "标题不能超过 100 个字符")
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = str.Slug(title, true)
	} else {
		slug = str.NormalizeSlug(slug)
	}
	for i := 0; i < 100; i++ {
		candidate := slug
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", slug, i+1)
		}
		exists, err := s.repo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return nil, err
		}
		if !exists {
			slug = candidate
			break
		}
	}
	status := model.PageStatus(in.Status)
	switch status {
	case model.PageStatusPublished, model.PageStatusArchived:
	case model.PageStatusDraft:
		// 历史草稿或旧页面提交的 draft 转为归档，避免意外发布。
		status = model.PageStatusArchived
	default:
		status = model.PageStatusPublished
	}
	return &model.Page{
		Title:          title,
		Slug:           slug,
		Content:        SanitizeHTML(in.Content),
		Status:         status,
		Sort:           in.Sort,
		SeoTitle:       SanitizeText(in.SeoTitle),
		SeoKeywords:    SanitizeText(in.SeoKeywords),
		SeoDescription: SanitizeText(in.SeoDescription),
	}, nil
}
