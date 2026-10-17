package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
)

// CategoryHandler 处理前台分类页。
type CategoryHandler struct{ *Base }

// NewCategoryHandler 创建分类 handler。
func NewCategoryHandler(base *Base) *CategoryHandler { return &CategoryHandler{Base: base} }

// Show 渲染分类页。
func (h *CategoryHandler) Show(c *gin.Context) {
	page := pageParam(c)
	result, err := h.Front.Category(c.Request.Context(), c.Param("slug"), page, 9)
	if err != nil {
		if err == apperr.ErrNotFound {
			h.NotFound(c, "分类不存在")
			return
		}
		h.ServerError(c, "加载分类失败")
		return
	}
	title := result.Category.SeoTitle
	if title == "" {
		title = result.Category.Name
	}
	description := result.Category.SeoDescription
	if description == "" {
		description = result.Category.Description
	}
	data := h.Data(c, title, description)
	data["Category"] = result.Category
	data["Children"] = result.Children
	data["Articles"] = result.Articles
	data["Total"] = result.Total
	data["Pager"] = pagination.NewView(result.Page, result.TotalPages, func(p int) string {
		if p <= 1 {
			return "/category/" + result.Category.Slug
		}
		return fmt.Sprintf("/category/%s/page/%d", result.Category.Slug, p)
	})
	c.HTML(http.StatusOK, "web/category.html", data)
}
