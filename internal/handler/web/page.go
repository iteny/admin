package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/apperr"
)

// PageHandler 处理前台单页。
type PageHandler struct{ *Base }

// NewPageHandler 创建单页 handler。
func NewPageHandler(base *Base) *PageHandler { return &PageHandler{Base: base} }

// Show 渲染单页。
func (h *PageHandler) Show(c *gin.Context) {
	page, err := h.Front.Page(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if err == apperr.ErrNotFound {
			h.NotFound(c, "页面不存在")
			return
		}
		h.ServerError(c, "加载页面失败")
		return
	}
	title := page.SeoTitle
	if title == "" {
		title = page.Title
	}
	description := page.SeoDescription
	if description == "" {
		description = page.Title
	}
	data := h.Data(c, title, description)
	data["Page"] = page
	c.HTML(http.StatusOK, "web/page.html", data)
}
