package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cms/internal/model"
	"cms/internal/pkg/response"
	"cms/internal/repository"
	"cms/internal/service"
)

// AccessLog 记录请求，并把后台写操作写入 operation_logs。
func AccessLog(logger *slog.Logger, logs *repository.LogRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		elapsed := time.Since(start)

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/uploads/") {
			logger.Debug("request", "method", c.Request.Method, "path", path, "status", c.Writer.Status(), "duration", elapsed)
			return
		}
		logger.Info("request", "method", c.Request.Method, "path", path, "status", c.Writer.Status(), "ip", c.ClientIP(), "duration", elapsed)

		if !strings.HasPrefix(path, "/admin") || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			return
		}
		user := CurrentUser(c)
		entry := &model.OperationLog{
			Module:     moduleOf(path),
			Action:     c.Request.Method + " " + path,
			Method:     c.Request.Method,
			Path:       path,
			IP:         c.ClientIP(),
			UserAgent:  truncate(c.Request.UserAgent(), 255),
			Request:    sanitizedRequest(c),
			StatusCode: c.Writer.Status(),
			DurationMs: int(elapsed.Milliseconds()),
		}
		if user != nil {
			entry.UserID = user.ID
			entry.Username = user.Username
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := logs.Create(ctx, entry); err != nil {
			logger.Warn("write operation log failed", "error", err)
		}
	}
}

// InjectSiteSettings 把站点设置注入每个请求的 context，供前台/后台布局使用。
func InjectSiteSettings(settings *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		all := settings.All()
		c.Set(SiteKey, all)
		name := settings.Get("site_name", "Tabler CMS")
		c.Set(SiteNameKey, name)
		c.Next()
	}
}

// Recovery 捕获 panic 并渲染 500 页面。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", "error", fmt.Sprint(rec), "path", c.Request.URL.Path, "stack", stack())
				if c.Writer.Written() {
					c.Abort()
					return
				}
				if isAJAX(c) {
					response.Fail(c, http.StatusInternalServerError, 50000, "服务器内部错误")
					return
				}
				c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
					"Title":   "服务器错误",
					"Message": "服务器处理请求时发生错误，请稍后重试。",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func moduleOf(path string) string {
	trimmed := strings.TrimPrefix(path, "/admin/")
	if trimmed == path {
		return "dashboard"
	}
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		trimmed = trimmed[:i]
	}
	if trimmed == "" {
		return "dashboard"
	}
	return trimmed
}

func sanitizedRequest(c *gin.Context) string {
	ct := c.ContentType()
	if strings.Contains(ct, "application/json") {
		return "[json body omitted]"
	}
	if strings.Contains(ct, "multipart/form-data") {
		return "[multipart body omitted]"
	}
	if err := c.Request.ParseForm(); err != nil {
		return ""
	}
	parts := make([]string, 0, len(c.Request.PostForm))
	for key, values := range c.Request.PostForm {
		if isSensitive(key) {
			parts = append(parts, key+"=[redacted]")
			continue
		}
		value := strings.Join(values, ",")
		parts = append(parts, key+"="+truncate(value, 120))
	}
	out := strings.Join(parts, "&")
	return truncate(out, 2000)
}

func isSensitive(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") || strings.Contains(k, "token") || strings.Contains(k, "csrf") || strings.Contains(k, "secret")
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
