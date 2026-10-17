package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/service"
)

// DashboardHandler 处理仪表盘。
type DashboardHandler struct {
	*Base
	Stats *service.DashboardService
}

// NewDashboardHandler 创建仪表盘 handler。
func NewDashboardHandler(base *Base, stats *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{Base: base, Stats: stats}
}

// Index 渲染仪表盘。
func (h *DashboardHandler) Index(c *gin.Context) {
	stats, err := h.Stats.Stats(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin")
		return
	}
	data := h.Data(c, "仪表盘", "概览", "dashboard")
	data["Stats"] = stats
	data["TrendJSON"] = stats.Trend
	c.HTML(http.StatusOK, "admin/dashboard/index.html", data)
}

// StatsJSON 返回近 7 日趋势。
func (h *DashboardHandler) StatsJSON(c *gin.Context) {
	stats, err := h.Stats.Stats(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin")
		return
	}
	h.JSON(c, gin.H{"trend": stats.Trend, "articles": stats.Articles, "views": stats.Views})
}
