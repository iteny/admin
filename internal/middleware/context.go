// Package middleware 提供后台认证、权限、CSRF、日志等 Gin 中间件。
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"cms/internal/model"
	"cms/internal/pkg/flash"
	"cms/internal/pkg/response"
)

const (
	// UserKey 是当前登录用户在 gin.Context 中的键。
	UserKey = "current_user"
	// PermissionsKey 是权限集合键。
	PermissionsKey = "permissions"
	// CSRFTokenKey 是 CSRF token 键。
	CSRFTokenKey = "csrf_token"
	// FlashKey 是 flash 消息键。
	FlashKey = "flash_messages"
	// SiteKey 是站点设置键。
	SiteKey = "site_settings"
	// SiteNameKey 是站点名称键。
	SiteNameKey = "site_name"
)

const (
	sessionUserID = "uid"
	sessionCSRF   = "csrf_token"
	flashSession  = "cms_flash"
)

// CurrentUser 返回当前登录用户。
func CurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get(UserKey); ok {
		if user, ok := v.(*model.User); ok {
			return user
		}
	}
	return nil
}

// Permissions 返回当前用户权限集合。
func Permissions(c *gin.Context) map[string]bool {
	if v, ok := c.Get(PermissionsKey); ok {
		if perms, ok := v.(map[string]bool); ok {
			return perms
		}
	}
	return map[string]bool{}
}

// Token 返回当前会话的 CSRF token。
func Token(c *gin.Context) string {
	if v, ok := c.Get(CSRFTokenKey); ok {
		if token, ok := v.(string); ok {
			return token
		}
	}
	return ""
}

// Messages 返回本次请求的 flash 消息。
func Messages(c *gin.Context) []flash.Message {
	if v, ok := c.Get(FlashKey); ok {
		if list, ok := v.([]flash.Message); ok {
			return list
		}
	}
	return nil
}

// SiteSettings 返回站点设置快照。
func SiteSettings(c *gin.Context) map[string]string {
	if v, ok := c.Get(SiteKey); ok {
		if m, ok := v.(map[string]string); ok {
			return m
		}
	}
	return map[string]string{}
}

// SiteName 返回站点名称。
func SiteName(c *gin.Context) string {
	if v, ok := c.Get(SiteNameKey); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "Tabler CMS"
}

// SetUser 把登录用户写入 context。
func SetUser(c *gin.Context, user *model.User, perms map[string]bool) {
	c.Set(UserKey, user)
	c.Set(PermissionsKey, perms)
}

// SessionSetUser 登录成功后写入 session。
func SessionSetUser(c *gin.Context, user *model.User) error {
	s := sessions.Default(c)
	s.Set(sessionUserID, user.ID)
	return s.Save()
}

// SessionUserID 读取 session 中的用户 ID。
func SessionUserID(c *gin.Context) uint64 {
	s := sessions.Default(c)
	switch v := s.Get(sessionUserID).(type) {
	case uint64:
		return v
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	default:
		return 0
	}
}

// SessionClear 清空登录态。
func SessionClear(c *gin.Context) error {
	s := sessions.Default(c)
	s.Clear()
	s.Options(sessions.Options{Path: "/", MaxAge: -1, HttpOnly: true})
	return s.Save()
}

// FlashMiddleware 把 session 中的一次性消息搬到 context。
func FlashMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		if v := s.Get(flashSession); v != nil {
			if list, ok := v.([]flash.Message); ok {
				c.Set(FlashKey, list)
			}
			s.Delete(flashSession)
			_ = s.Save()
		}
		c.Next()
	}
}

// CSRF 校验全站表单与 AJAX 的 token。
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") || strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Next()
			return
		}
		s := sessions.Default(c)
		token, _ := s.Get(sessionCSRF).(string)
		if token == "" {
			token = randomToken(32)
			s.Set(sessionCSRF, token)
			_ = s.Save()
		}
		c.Set(CSRFTokenKey, token)

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			c.Next()
			return
		}
		provided := c.GetHeader("X-CSRF-Token")
		if provided == "" && c.ContentType() != "multipart/form-data" {
			provided = c.PostForm("_csrf")
		}
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			if isAJAX(c) {
				response.Fail(c, http.StatusTeapot, 41900, "CSRF token 已失效，请刷新页面重试")
			} else {
				c.HTML(http.StatusTeapot, "errors/419.html", gin.H{"Title": "页面已过期", "Message": "安全令牌已失效，请刷新页面重试。"})
				c.Abort()
			}
			return
		}
		c.Next()
	}
}

// MethodOverride 允许 POST 表单通过 _method 覆盖为 PUT/PATCH/DELETE。
func MethodOverride() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只解析 urlencoded 表单；multipart 上传在 CSRF/MethodOverride 阶段
		// 解析会把整个文件读进内存，必须等到 handler 设置 MaxBytesReader 后再读。
		if c.Request.Method == http.MethodPost && c.ContentType() == "application/x-www-form-urlencoded" {
			if method := strings.ToUpper(strings.TrimSpace(c.PostForm("_method"))); method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
				c.Request.Method = method
			}
		}
		c.Next()
	}
}

func randomToken(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-csrf-token"
	}
	return hex.EncodeToString(buf)
}

func isAJAX(c *gin.Context) bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(c.GetHeader("Accept"), "application/json") ||
		strings.HasPrefix(c.Request.URL.Path, "/admin/api/")
}
