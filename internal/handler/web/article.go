package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/apperr"
)

// ArticleHandler 处理前台文章页。
type ArticleHandler struct{ *Base }

// NewArticleHandler 创建文章 handler。
func NewArticleHandler(base *Base) *ArticleHandler { return &ArticleHandler{Base: base} }

// Show 渲染文章详情。
func (h *ArticleHandler) Show(c *gin.Context) {
	detail, err := h.Front.Article(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if err == apperr.ErrNotFound {
			h.NotFound(c, "文章不存在或尚未发布")
			return
		}
		h.ServerError(c, "加载文章失败")
		return
	}
	title := detail.Article.SeoTitle
	if title == "" {
		title = detail.Article.Title
	}
	description := detail.Article.SeoDescription
	if description == "" {
		description = detail.Article.Summary
	}
	data := h.Data(c, title, description)
	data["Article"] = detail.Article
	data["Prev"] = detail.Prev
	data["Next"] = detail.Next
	data["Related"] = detail.Related
	data["Category"] = detail.Category
	data["Canonical"] = h.Config.App.BaseURL + "/article/" + detail.Article.Slug
	published := detail.Article.CreatedAt
	if detail.Article.PublishedAt != nil {
		published = *detail.Article.PublishedAt
	}
	data["JSONLD"] = map[string]any{
		"@context":         "https://schema.org",
		"@type":            "Article",
		"headline":         detail.Article.Title,
		"description":      detail.Article.SeoDescription,
		"image":            detail.Article.CoverURL,
		"datePublished":    published.Format(time.RFC3339),
		"dateModified":     detail.Article.UpdatedAt.Format(time.RFC3339),
		"author":           map[string]any{"@type": "Person", "name": detail.Article.Author.Nickname},
		"publisher":        map[string]any{"@type": "Organization", "name": h.Settings.Get("site_name", "Tabler CMS")},
		"mainEntityOfPage": h.Config.App.BaseURL + "/article/" + detail.Article.Slug,
	}
	c.HTML(http.StatusOK, "web/article.html", data)
}
