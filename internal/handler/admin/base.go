package admin

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/config"
	"cms/internal/middleware"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/flash"
	"cms/internal/pkg/response"
	"cms/internal/service"
)

// Base 提供后台 handler 共用依赖和视图数据。
type Base struct {
	Config   *config.Config
	Settings *service.SettingService
	Menus    *service.MenuService
	Logger   *slog.Logger
	Loc      *time.Location
	Version  string
}

// NewBase 创建后台基础 handler。
func NewBase(cfg *config.Config, settings *service.SettingService, menus *service.MenuService, logger *slog.Logger, loc *time.Location, version string) *Base {
	return &Base{Config: cfg, Settings: settings, Menus: menus, Logger: logger, Loc: loc, Version: version}
}

// Data 构造后台布局通用数据。
func (b *Base) Data(c *gin.Context, title, pretitle, active string) gin.H {
	site := middleware.SiteSettings(c)
	return gin.H{
		"Title":       title,
		"Pretitle":    pretitle,
		"Active":      active,
		"SiteName":    middleware.SiteName(c),
		"Site":        site,
		"User":        middleware.CurrentUser(c),
		"Perms":       middleware.Permissions(c),
		"CSRFToken":   middleware.Token(c),
		"Flash":       middleware.Messages(c),
		"Menus":       middleware.AdminMenuTree(c),
		"CurrentPath": c.Request.URL.Path,
		"Query":       c.Request.URL.Query(),
		"Version":     b.Version,
		"Debug":       b.Config.App.Debug,
		"Year":        time.Now().In(b.Loc).Year(),
	}
}

// ErrorMessage 把领域错误转成用户可读消息。
func (b *Base) ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var ve *apperr.ValidationError
	switch {
	case errors.As(err, &ve):
		return ve.Error()
	case errors.Is(err, apperr.ErrNotFound):
		return "资源不存在或已被删除"
	case errors.Is(err, apperr.ErrForbidden):
		return "没有权限执行此操作"
	case errors.Is(err, apperr.ErrConflict):
		return "资源冲突，请检查输入"
	default:
		return err.Error()
	}
}

// Back 返回安全的返回地址。
func (b *Base) Back(c *gin.Context, fallback string) string {
	if ref := c.GetHeader("Referer"); ref != "" {
		if u, err := url.Parse(ref); err == nil {
			target := u.EscapedPath()
			if u.RawQuery != "" {
				target += "?" + u.RawQuery
			}
			return middleware.SafeRedirect(target, fallback)
		}
	}
	return fallback
}

// Fail 把错误写回页面：AJAX 返回 JSON，普通表单写 flash 并跳回。
func (b *Base) Fail(c *gin.Context, err error, back string) {
	msg := b.ErrorMessage(err)
	if b.Config.App.Debug && err != nil {
		b.Logger.Warn("admin request failed", "error", err, "path", c.Request.URL.Path)
	}
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" || c.GetHeader("Accept") == "application/json" {
		status := http.StatusBadRequest
		code := 40000
		switch {
		case errors.Is(err, apperr.ErrNotFound):
			status, code = http.StatusNotFound, 40400
		case errors.Is(err, apperr.ErrForbidden):
			status, code = http.StatusForbidden, 40300
		case errors.Is(err, apperr.ErrUnauthorized):
			status, code = http.StatusUnauthorized, 40100
		}
		response.Fail(c, status, code, msg)
		return
	}
	flash.Error(c, msg)
	c.Redirect(http.StatusFound, back)
	c.Abort()
}

// Success 写入成功提示并跳转。
func (b *Base) Success(c *gin.Context, message, target string) {
	flash.Success(c, message)
	c.Redirect(http.StatusFound, target)
	c.Abort()
}

// JSON 返回成功 JSON。
func (b *Base) JSON(c *gin.Context, data any) {
	response.OK(c, data)
}

// Bind 绑定表单，失败时返回错误信息。
func (b *Base) Bind(c *gin.Context, obj any) error {
	if err := c.ShouldBind(obj); err != nil {
		return apperr.NewValidation("form", "表单填写有误，请检查必填项和格式")
	}
	return nil
}
