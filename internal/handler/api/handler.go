// Package api 提供 Android / 移动端使用的只读 JSON API。
package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/config"
	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
	"cms/internal/pkg/response"
	"cms/internal/repository"
	"cms/internal/service"
)

// Handler 处理公开 API 请求。
type Handler struct {
	Config      *config.Config
	Settings    *service.SettingService
	Front       *service.FrontService
	ArticleSvc  *service.ArticleService
	CategorySvc *service.CategoryService
	TagSvc      *service.TagService
	Logger      *slog.Logger
	Loc         *time.Location
	Version     string
}

// New 创建 API handler。
func New(cfg *config.Config, settings *service.SettingService, front *service.FrontService,
	articles *service.ArticleService, categories *service.CategoryService, tags *service.TagService,
	logger *slog.Logger, loc *time.Location, version string) *Handler {
	return &Handler{
		Config: cfg, Settings: settings, Front: front, ArticleSvc: articles,
		CategorySvc: categories, TagSvc: tags, Logger: logger, Loc: loc, Version: version,
	}
}

// CategoryBrief 是文章中的分类摘要。
type CategoryBrief struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// TagBrief 是文章中的标签摘要。
type TagBrief struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// AuthorBrief 是文章作者摘要。
type AuthorBrief struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}

// ArticleCard 是列表中的文章摘要。
type ArticleCard struct {
	ID          uint64         `json:"id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Summary     string         `json:"summary"`
	CoverURL    string         `json:"cover_url,omitempty"`
	Category    *CategoryBrief `json:"category,omitempty"`
	Tags        []TagBrief     `json:"tags"`
	Author      *AuthorBrief   `json:"author,omitempty"`
	IsTop       bool           `json:"is_top"`
	IsRecommend bool           `json:"is_recommend"`
	ViewCount   int            `json:"view_count"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
	URL         string         `json:"url"`
}

// ArticleWithContent 是可选返回正文的列表项。
type ArticleWithContent struct {
	ArticleCard
	ContentHTML    string `json:"content_html"`
	SubContentHTML string `json:"sub_content_html"`
	ContentBaseURL string `json:"content_base_url"`
}

// ArticleDetail 是文章详情。
type ArticleDetail struct {
	ArticleCard
	ContentHTML    string        `json:"content_html"`
	SubContentHTML string        `json:"sub_content_html"`
	ContentType    string        `json:"content_type"`
	ContentBaseURL string        `json:"content_base_url"`
	SeoTitle       string        `json:"seo_title,omitempty"`
	SeoKeywords    string        `json:"seo_keywords,omitempty"`
	SeoDescription string        `json:"seo_description,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Prev           *ArticleCard  `json:"prev,omitempty"`
	Next           *ArticleCard  `json:"next,omitempty"`
	Related        []ArticleCard `json:"related"`
}

// Category 是分类树节点。
type Category struct {
	ID           uint64     `json:"id"`
	ParentID     uint64     `json:"parent_id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  string     `json:"description,omitempty"`
	CoverURL     string     `json:"cover_url,omitempty"`
	Sort         int        `json:"sort"`
	ArticleCount int64      `json:"article_count"`
	Children     []Category `json:"children"`
}

// Tag 是标签。
type Tag struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	UseCount int    `json:"use_count"`
}

// Site 返回站点信息，供 App 启动时读取。
func (h *Handler) Site(c *gin.Context) {
	response.OK(c, gin.H{
		"name":        h.Settings.Get("site_name", "Tabler CMS"),
		"description": h.Settings.Get("seo_description", ""),
		"base_url":    h.baseURL(c),
		"version":     h.Version,
	})
}

// Home 返回 App 首页聚合数据。
func (h *Handler) Home(c *gin.Context) {
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "page_size", 10)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 20 {
		pageSize = 20
	}
	data, err := h.Front.Home(c.Request.Context(), page, pageSize)
	if err != nil {
		h.serverError(c, err)
		return
	}
	base := h.baseURL(c)
	response.OK(c, gin.H{
		"recommended": h.cards(data.Recommended, base),
		"latest":      pagination.New(h.cards(data.Latest, base), data.Total, data.Page, data.PageSize),
		"top":         h.cards(data.Top, base),
		"hot":         h.cards(data.Hot, base),
		"categories":  h.categoryTree(data.Categories, base),
		"tags":        h.tagList(data.Tags),
	})
}

// Articles 返回已发布文章列表。
//
// 查询参数：page、page_size、category、category_id、tag、tag_id、q/keyword、
// sort(latest|oldest|views|updated)、featured、top。
func (h *Handler) Articles(c *gin.Context) {
	ctx := c.Request.Context()
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "page_size", 10)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	filter := repository.ArticleFilter{
		Query:  pagination.Query{Page: page, PageSize: pageSize, Keyword: firstQuery(c, "q", "keyword"), Sort: c.Query("sort")},
		Status: string(model.ArticleStatusPublished),
	}

	if raw := strings.TrimSpace(c.Query("category")); raw != "" {
		category, err := h.CategorySvc.GetBySlug(ctx, raw)
		if err != nil {
			h.notFoundOrServer(c, err, "分类不存在")
			return
		}
		ids, err := h.CategorySvc.DescendantIDs(ctx, category.ID)
		if err != nil {
			h.serverError(c, err)
			return
		}
		filter.CategoryIDs = ids
	} else if raw := strings.TrimSpace(c.Query("category_id")); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, 40000, "category_id 无效")
			return
		}
		category, err := h.CategorySvc.Get(ctx, id)
		if err != nil {
			h.notFoundOrServer(c, err, "分类不存在")
			return
		}
		ids, err := h.CategorySvc.DescendantIDs(ctx, category.ID)
		if err != nil {
			h.serverError(c, err)
			return
		}
		filter.CategoryIDs = ids
	}

	if raw := strings.TrimSpace(c.Query("tag")); raw != "" {
		tag, err := h.TagSvc.GetBySlug(ctx, raw)
		if err != nil {
			h.notFoundOrServer(c, err, "标签不存在")
			return
		}
		filter.TagID = tag.ID
	} else if raw := strings.TrimSpace(c.Query("tag_id")); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, 40000, "tag_id 无效")
			return
		}
		tag, err := h.TagSvc.Get(ctx, id)
		if err != nil {
			h.notFoundOrServer(c, err, "标签不存在")
			return
		}
		filter.TagID = tag.ID
	}

	if value, ok := boolQuery(c, "featured"); ok {
		filter.IsRecommend = &value
	}
	if value, ok := boolQuery(c, "top"); ok {
		filter.IsTop = &value
	}

	list, total, err := h.ArticleSvc.List(ctx, filter)
	if err != nil {
		h.serverError(c, err)
		return
	}
	base := h.baseURL(c)
	cards := h.cards(list, base)
	if include, _ := boolQuery(c, "include_content"); include {
		items := make([]ArticleWithContent, 0, len(cards))
		for i := range cards {
			items = append(items, ArticleWithContent{
				ArticleCard:    cards[i],
				ContentHTML:    list[i].Content,
				SubContentHTML: list[i].SubContent,
				ContentBaseURL: base,
			})
		}
		response.OK(c, pagination.New(items, total, page, pageSize))
		return
	}
	response.OK(c, pagination.New(cards, total, page, pageSize))
}

// Article 返回文章详情，并增加浏览量。
func (h *Handler) Article(c *gin.Context) {
	h.renderArticle(c, c.Param("slug"))
}

// ArticleByID 按文章 ID 返回详情。
func (h *Handler) ArticleByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, 40000, "文章 ID 无效")
		return
	}
	article, err := h.ArticleSvc.Get(c.Request.Context(), id)
	if err != nil {
		h.notFoundOrServer(c, err, "文章不存在或未发布")
		return
	}
	if article.Status != model.ArticleStatusPublished {
		response.Fail(c, http.StatusNotFound, 40400, "文章不存在或未发布")
		return
	}
	h.renderArticle(c, article.Slug)
}

func (h *Handler) renderArticle(c *gin.Context, slug string) {
	detail, err := h.Front.Article(c.Request.Context(), slug)
	if err != nil {
		h.notFoundOrServer(c, err, "文章不存在或未发布")
		return
	}
	base := h.baseURL(c)
	item := detail.Article
	out := ArticleDetail{
		ArticleCard:    h.card(*item, base),
		ContentHTML:    item.Content,
		SubContentHTML: item.SubContent,
		ContentType:    string(item.ContentType),
		ContentBaseURL: base,
		SeoTitle:       item.SeoTitle,
		SeoKeywords:    item.SeoKeywords,
		SeoDescription: item.SeoDescription,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		Related:        h.cards(detail.Related, base),
	}
	if detail.Prev != nil {
		card := h.card(*detail.Prev, base)
		out.Prev = &card
	}
	if detail.Next != nil {
		card := h.card(*detail.Next, base)
		out.Next = &card
	}
	response.OK(c, out)
}

// Categories 返回启用分类树。
func (h *Handler) Categories(c *gin.Context) {
	nodes, err := h.CategorySvc.Tree(c.Request.Context(), true)
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.OK(c, gin.H{"list": h.categoryTree(nodes, h.baseURL(c))})
}

// Tags 返回标签列表。
func (h *Handler) Tags(c *gin.Context) {
	tags, err := h.TagSvc.List(c.Request.Context())
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.OK(c, gin.H{"list": h.tagList(tags)})
}

func (h *Handler) cards(list []model.Article, base string) []ArticleCard {
	out := make([]ArticleCard, 0, len(list))
	for i := range list {
		out = append(out, h.card(list[i], base))
	}
	return out
}

func (h *Handler) card(item model.Article, base string) ArticleCard {
	card := ArticleCard{
		ID:          item.ID,
		Title:       item.Title,
		Slug:        item.Slug,
		Summary:     item.Summary,
		CoverURL:    absoluteURL(base, item.CoverURL),
		Tags:        tagBriefs(item.Tags),
		IsTop:       item.IsTop,
		IsRecommend: item.IsRecommend,
		ViewCount:   item.ViewCount,
		PublishedAt: item.PublishedAt,
		URL:         absoluteURL(base, "/article/"+item.Slug),
	}
	if item.Category.ID > 0 {
		card.Category = &CategoryBrief{ID: item.Category.ID, Name: item.Category.Name, Slug: item.Category.Slug}
	}
	if item.Author.ID > 0 {
		nickname := item.Author.Nickname
		if nickname == "" {
			nickname = item.Author.Username
		}
		card.Author = &AuthorBrief{ID: item.Author.ID, Nickname: nickname, Avatar: absoluteURL(base, item.Author.Avatar)}
	}
	return card
}

func tagBriefs(tags []model.Tag) []TagBrief {
	out := make([]TagBrief, 0, len(tags))
	for _, tag := range tags {
		out = append(out, TagBrief{ID: tag.ID, Name: tag.Name, Slug: tag.Slug})
	}
	return out
}

func (h *Handler) tagList(tags []model.Tag) []Tag {
	out := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, Tag{ID: tag.ID, Name: tag.Name, Slug: tag.Slug, UseCount: tag.UseCount})
	}
	return out
}

func (h *Handler) categoryTree(nodes []*service.CategoryNode, base string) []Category {
	out := make([]Category, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, Category{
			ID:           node.Category.ID,
			ParentID:     node.Category.ParentID,
			Name:         node.Category.Name,
			Slug:         node.Category.Slug,
			Description:  node.Category.Description,
			CoverURL:     absoluteURL(base, node.Category.CoverURL),
			Sort:         node.Category.Sort,
			ArticleCount: node.ArticleCount,
			Children:     h.categoryTree(node.Children, base),
		})
	}
	return out
}

func (h *Handler) baseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if raw := c.GetHeader("X-Forwarded-Proto"); raw != "" {
		if value := strings.TrimSpace(strings.Split(raw, ",")[0]); value != "" {
			scheme = value
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return strings.TrimRight(h.Config.App.BaseURL, "/")
	}
	return scheme + "://" + host
}

func absoluteURL(base, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "//") {
		return value
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(value, "/")
}

func (h *Handler) notFoundOrServer(c *gin.Context, err error, message string) {
	if errors.Is(err, apperr.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, 40400, message)
		return
	}
	h.serverError(c, err)
}

func (h *Handler) serverError(c *gin.Context, err error) {
	if h.Logger != nil {
		h.Logger.Error("api request failed", "path", c.Request.URL.Path, "error", err)
	}
	response.Fail(c, http.StatusInternalServerError, 50000, "服务器内部错误")
}

func intQuery(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return value
}

func boolQuery(c *gin.Context, key string) (bool, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return false, false
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func firstQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return ""
}
