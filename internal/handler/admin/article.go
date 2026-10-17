package admin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
	"cms/internal/repository"
	"cms/internal/service"
)

// ArticleHandler 处理文章 CRUD。
type ArticleHandler struct {
	*Base
	Articles   *service.ArticleService
	Categories *service.CategoryService
	Tags       *service.TagService
}

// NewArticleHandler 创建文章 handler。
func NewArticleHandler(base *Base, articles *service.ArticleService, categories *service.CategoryService, tags *service.TagService) *ArticleHandler {
	return &ArticleHandler{Base: base, Articles: articles, Categories: categories, Tags: tags}
}

type articleForm struct {
	Title          string
	Slug           string
	Summary        string
	Content        string
	SubContent     string
	ContentType    string
	CoverURL       string
	CategoryID     uint64
	Status         string
	PublishedAt    string
	IsTop          string
	IsRecommend    string
	SeoTitle       string
	SeoKeywords    string
	SeoDescription string
}

// Index 渲染文章列表。
func (h *ArticleHandler) Index(c *gin.Context) {
	q := pagination.Query{
		Page:     intParam(c, "page", 1),
		PageSize: intParam(c, "page_size", 15),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Sort:     c.Query("sort"),
	}
	status := c.Query("status")
	categoryID := uint64Param(c, "category_id")
	tagID := uint64Param(c, "tag_id")
	list, total, err := h.Articles.List(c.Request.Context(), repository.ArticleFilter{
		Query: q, Status: status, CategoryID: categoryID, TagID: tagID,
	})
	if err != nil {
		h.Fail(c, err, "/admin/articles")
		return
	}
	categories, _ := h.Categories.Options(c.Request.Context())
	tags, _ := h.Tags.List(c.Request.Context())
	data := h.Data(c, "文章列表", "内容管理", "article")
	data["Data"] = pagination.New(list, total, q.Page, q.PageSize)
	data["Categories"] = categories
	data["Tags"] = tags
	data["Status"] = status
	data["CategoryID"] = categoryID
	data["TagID"] = tagID
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/articles"
	c.HTML(http.StatusOK, "admin/article/index.html", data)
}

// CreatePage 渲染新增文章页。
func (h *ArticleHandler) CreatePage(c *gin.Context) {
	h.renderForm(c, nil, nil, "")
}

// Create 创建文章。
func (h *ArticleHandler) Create(c *gin.Context) {
	form, tagIDs := bindArticleForm(c)
	user := h.currentUser(c)
	if user == nil {
		h.Fail(c, apperr.ErrUnauthorized, "/admin/articles")
		return
	}
	article, err := h.Articles.Create(c.Request.Context(), form.toInput(tagIDs, h.Loc), user.ID)
	if err != nil {
		h.renderForm(c, form.toArticle(tagIDs), tagIDs, h.ErrorMessage(err))
		return
	}
	h.Success(c, "文章已创建", "/admin/articles/"+strconv.FormatUint(article.ID, 10)+"/edit")
}

// EditPage 渲染编辑页。
func (h *ArticleHandler) EditPage(c *gin.Context) {
	id := uint64Param(c, "id")
	article, err := h.Articles.Get(c.Request.Context(), id)
	if err != nil {
		h.Fail(c, err, "/admin/articles")
		return
	}
	selected := make(map[uint64]bool, len(article.Tags))
	for _, tag := range article.Tags {
		selected[tag.ID] = true
	}
	h.renderArticle(c, article, selected, "")
}

// Update 更新文章。
func (h *ArticleHandler) Update(c *gin.Context) {
	id := uint64Param(c, "id")
	form, tagIDs := bindArticleForm(c)
	user := h.currentUser(c)
	if user == nil {
		h.Fail(c, apperr.ErrUnauthorized, "/admin/articles")
		return
	}
	article, err := h.Articles.Update(c.Request.Context(), id, form.toInput(tagIDs, h.Loc), user.ID)
	if err != nil {
		existing, _ := h.Articles.Get(c.Request.Context(), id)
		if existing == nil {
			existing = &model.Article{}
		}
		existing.Title = form.Title
		existing.Slug = form.Slug
		existing.Summary = form.Summary
		existing.Content = form.Content
		existing.SubContent = form.SubContent
		existing.ContentType = model.ContentType(form.ContentType)
		existing.CoverURL = form.CoverURL
		existing.CategoryID = form.CategoryID
		existing.Status = model.ArticleStatus(form.Status)
		existing.IsTop = form.IsTop == "on"
		existing.IsRecommend = form.IsRecommend == "on"
		existing.SeoTitle = form.SeoTitle
		existing.SeoKeywords = form.SeoKeywords
		existing.SeoDescription = form.SeoDescription
		selected := make(map[uint64]bool)
		for _, id := range tagIDs {
			selected[id] = true
		}
		h.renderArticle(c, existing, selected, h.ErrorMessage(err))
		return
	}
	h.Success(c, "文章已保存", "/admin/articles/"+strconv.FormatUint(article.ID, 10)+"/edit")
}

// Delete 删除文章。
func (h *ArticleHandler) Delete(c *gin.Context) {
	id := uint64Param(c, "id")
	if err := h.Articles.Delete(c.Request.Context(), id); err != nil {
		h.Fail(c, err, "/admin/articles")
		return
	}
	h.Success(c, "文章已删除", "/admin/articles")
}

// Batch 批量操作。
func (h *ArticleHandler) Batch(c *gin.Context) {
	ids := parseIDs(c.PostFormArray("ids"))
	action := c.PostForm("batch_action")
	if err := h.Articles.Batch(c.Request.Context(), ids, action); err != nil {
		h.Fail(c, err, "/admin/articles")
		return
	}
	h.Success(c, "批量操作已完成", "/admin/articles")
}

// Preview 预览文章。
func (h *ArticleHandler) Preview(c *gin.Context) {
	article, err := h.Articles.Get(c.Request.Context(), uint64Param(c, "id"))
	if err != nil {
		h.Fail(c, err, "/admin/articles")
		return
	}
	data := h.Data(c, article.Title, "文章预览", "article")
	data["Article"] = article
	c.HTML(http.StatusOK, "admin/article/preview.html", data)
}

func (h *ArticleHandler) renderForm(c *gin.Context, article *model.Article, tagIDs []uint64, message string) {
	if article == nil {
		article = &model.Article{ContentType: model.ContentTypeHTML, Status: model.ArticleStatusPublished}
	}
	if article.Status == model.ArticleStatusDraft {
		// 兼容未迁移的历史草稿：编辑页按归档展示。
		article.Status = model.ArticleStatusArchived
	}
	selected := make(map[uint64]bool)
	for _, id := range tagIDs {
		selected[id] = true
	}
	h.renderArticle(c, article, selected, message)
}

func (h *ArticleHandler) renderArticle(c *gin.Context, article *model.Article, selected map[uint64]bool, message string) {
	categories, _ := h.Categories.Options(c.Request.Context())
	tags, _ := h.Tags.List(c.Request.Context())
	title := "新增文章"
	if article.ID > 0 {
		title = "编辑文章"
	}
	data := h.Data(c, title, "内容管理", "article")
	data["Article"] = article
	data["Categories"] = categories
	data["Tags"] = tags
	data["SelectedTags"] = selected
	data["Error"] = message
	data["IsEdit"] = article.ID > 0
	data["ActionURL"] = "/admin/articles"
	if article.ID > 0 {
		data["ActionURL"] = "/admin/articles/" + strconv.FormatUint(article.ID, 10)
	}
	c.HTML(http.StatusOK, "admin/article/form.html", data)
}

func (h *ArticleHandler) currentUser(c *gin.Context) *model.User {
	return middleware.CurrentUser(c)
}

func bindArticleForm(c *gin.Context) (articleForm, []uint64) {
	form := articleForm{
		Title:          strings.TrimSpace(c.PostForm("title")),
		Slug:           strings.TrimSpace(c.PostForm("slug")),
		Summary:        strings.TrimSpace(c.PostForm("summary")),
		Content:        c.PostForm("content"),
		SubContent:     c.PostForm("sub_content"),
		ContentType:    c.PostForm("content_type"),
		CoverURL:       strings.TrimSpace(c.PostForm("cover_url")),
		CategoryID:     uint64Param(c, "category_id"),
		Status:         c.PostForm("status"),
		PublishedAt:    c.PostForm("published_at"),
		IsTop:          c.PostForm("is_top"),
		IsRecommend:    c.PostForm("is_recommend"),
		SeoTitle:       c.PostForm("seo_title"),
		SeoKeywords:    c.PostForm("seo_keywords"),
		SeoDescription: c.PostForm("seo_description"),
	}
	return form, parseIDs(c.PostFormArray("tag_ids"))
}

func (f articleForm) toInput(tagIDs []uint64, loc *time.Location) service.ArticleInput {
	return service.ArticleInput{
		Title: f.Title, Slug: f.Slug, Summary: f.Summary, Content: f.Content, SubContent: f.SubContent,
		ContentType: f.ContentType, CoverURL: f.CoverURL, CategoryID: f.CategoryID,
		Status: f.Status, IsTop: f.IsTop == "on", IsRecommend: f.IsRecommend == "on",
		PublishedAt: parseDateTime(f.PublishedAt, loc), TagIDs: tagIDs,
		SeoTitle: f.SeoTitle, SeoKeywords: f.SeoKeywords, SeoDescription: f.SeoDescription,
	}
}

func (f articleForm) toArticle(tagIDs []uint64) *model.Article {
	a := &model.Article{
		Title: f.Title, Slug: f.Slug, Summary: f.Summary, Content: f.Content, SubContent: f.SubContent,
		ContentType: model.ContentType(f.ContentType), CoverURL: f.CoverURL,
		CategoryID: f.CategoryID, Status: model.ArticleStatus(f.Status),
		IsTop: f.IsTop == "on", IsRecommend: f.IsRecommend == "on",
		SeoTitle: f.SeoTitle, SeoKeywords: f.SeoKeywords, SeoDescription: f.SeoDescription,
	}
	for _, id := range tagIDs {
		a.Tags = append(a.Tags, model.Tag{BaseModel: model.BaseModel{ID: id}})
	}
	return a
}

func parseDateTime(value string, loc *time.Location) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}

func parseIDs(values []string) []uint64 {
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		if id, err := strconv.ParseUint(value, 10, 64); err == nil && id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func intParam(c *gin.Context, key string, def int) int {
	if value, err := strconv.Atoi(c.Query(key)); err == nil {
		return value
	}
	return def
}

func uint64Param(c *gin.Context, key string) uint64 {
	if value, err := strconv.ParseUint(c.Param(key), 10, 64); err == nil {
		return value
	}
	if value, err := strconv.ParseUint(c.Query(key), 10, 64); err == nil {
		return value
	}
	if value, err := strconv.ParseUint(c.PostForm(key), 10, 64); err == nil {
		return value
	}
	return 0
}
