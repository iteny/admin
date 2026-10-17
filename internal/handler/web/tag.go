package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
)

// TagHandler 处理前台标签页。
type TagHandler struct{ *Base }

// NewTagHandler 创建标签 handler。
func NewTagHandler(base *Base) *TagHandler { return &TagHandler{Base: base} }

// Show 渲染标签页。
func (h *TagHandler) Show(c *gin.Context) {
	page := pageParam(c)
	result, err := h.Front.Tag(c.Request.Context(), c.Param("slug"), page, 9)
	if err != nil {
		if err == apperr.ErrNotFound {
			h.NotFound(c, "标签不存在")
			return
		}
		h.ServerError(c, "加载标签失败")
		return
	}
	data := h.Data(c, "标签："+result.Tag.Name, "包含标签 "+result.Tag.Name+" 的文章")
	data["Tag"] = result.Tag
	data["Articles"] = result.Articles
	data["Total"] = result.Total
	data["Pager"] = pagination.NewView(result.Page, result.TotalPages, func(p int) string {
		if p <= 1 {
			return "/tag/" + result.Tag.Slug
		}
		return fmt.Sprintf("/tag/%s/page/%d", result.Tag.Slug, p)
	})
	c.HTML(http.StatusOK, "web/category.html", data)
}
