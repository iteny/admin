package web

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/config"
	"cms/internal/middleware"
	"cms/internal/service"
)

// Base 提供前台 handler 共用依赖。
type Base struct {
	Config   *config.Config
	Settings *service.SettingService
	Front    *service.FrontService
	Logger   *slog.Logger
	Loc      *time.Location
	Version  string
}

// NewBase 创建前台基础 handler。
func NewBase(cfg *config.Config, settings *service.SettingService, front *service.FrontService, logger *slog.Logger, loc *time.Location, version string) *Base {
	return &Base{Config: cfg, Settings: settings, Front: front, Logger: logger, Loc: loc, Version: version}
}

// Data 构造前台布局通用数据。
func (b *Base) Data(c *gin.Context, title, description string) gin.H {
	site := middleware.SiteSettings(c)
	if title == "" {
		title = b.Settings.Get("seo_title", middleware.SiteName(c))
	}
	if description == "" {
		description = b.Settings.Get("seo_description", "")
	}
	return gin.H{
		"Title":         title,
		"Description":   description,
		"SiteName":      middleware.SiteName(c),
		"Site":          site,
		"CSRFToken":     middleware.Token(c),
		"Flash":         middleware.Messages(c),
		"NavCategories": middleware.NavCategories(c),
		"NavPages":      middleware.NavPages(c),
		"CurrentPath":   c.Request.URL.Path,
		"Version":       b.Version,
		"Year":          time.Now().In(b.Loc).Year(),
	}
}

// NotFound 渲染前台 404。
func (b *Base) NotFound(c *gin.Context, message string) {
	data := b.Data(c, "页面不存在", message)
	data["Message"] = message
	c.HTML(http.StatusNotFound, "errors/404.html", data)
	c.Abort()
}

// ServerError 渲染前台 500。
func (b *Base) ServerError(c *gin.Context, message string) {
	data := b.Data(c, "服务器错误", message)
	data["Message"] = message
	c.HTML(http.StatusInternalServerError, "errors/500.html", data)
	c.Abort()
}

func pageParam(c *gin.Context) int {
	raw := c.Param("page")
	if raw == "" {
		raw = c.Query("page")
	}
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}
