package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/pkg/pagination"
)

// HomeHandler 处理前台首页。
type HomeHandler struct{ *Base }

// NewHomeHandler 创建首页 handler。
func NewHomeHandler(base *Base) *HomeHandler { return &HomeHandler{Base: base} }

// Index 渲染首页。
func (h *HomeHandler) Index(c *gin.Context) {
	page := pageParam(c)
	home, err := h.Front.Home(c.Request.Context(), page, 9)
	if err != nil {
		h.ServerError(c, "加载首页失败")
		return
	}
	data := h.Data(c, h.Settings.Get("seo_title", h.Settings.Get("site_name", "Tabler CMS")), h.Settings.Get("seo_description", ""))
	data["Home"] = home
	data["Pager"] = pagination.NewView(home.Page, home.TotalPages, func(p int) string {
		if p <= 1 {
			return "/"
		}
		return fmt.Sprintf("/page/%d", p)
	})
	c.HTML(http.StatusOK, "web/home.html", data)
}
