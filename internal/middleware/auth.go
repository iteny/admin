package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"cms/internal/pkg/response"
	"cms/internal/service"
)

// AuthRequired 要求后台登录，并把用户和权限注入 context。
func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := SessionUserID(c)
		if uid == 0 {
			redirectToLogin(c)
			return
		}
		user, err := auth.CurrentUser(c.Request.Context(), uid)
		if err != nil || user.Status == 0 {
			_ = SessionClear(c)
			redirectToLogin(c)
			return
		}
		SetUser(c, user, auth.Permissions(user))
		c.Next()
	}
}

// GuestOnly 让已登录用户跳过登录页。
func GuestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if SessionUserID(c) != 0 {
			c.Redirect(http.StatusFound, "/admin")
			c.Abort()
			return
		}
		c.Next()
	}
}

// Permission 要求指定权限点；超级管理员直接放行。
func Permission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := Permissions(c)
		if perms["*"] || perms[code] {
			c.Next()
			return
		}
		if isAJAX(c) {
			response.Fail(c, http.StatusForbidden, 40300, "没有权限执行此操作")
			return
		}
		c.HTML(http.StatusForbidden, "errors/403.html", gin.H{
			"Title":   "无权访问",
			"Message": "你的账号没有访问该页面的权限。",
		})
		c.Abort()
	}
}

// RequireAnyPermission 要求任意一个权限点。
func RequireAnyPermission(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := Permissions(c)
		if perms["*"] {
			c.Next()
			return
		}
		for _, code := range codes {
			if perms[code] {
				c.Next()
				return
			}
		}
		c.HTML(http.StatusForbidden, "errors/403.html", gin.H{"Title": "无权访问", "Message": "你的账号没有访问该页面的权限。"})
		c.Abort()
	}
}

func redirectToLogin(c *gin.Context) {
	if isAJAX(c) {
		response.Fail(c, http.StatusUnauthorized, 40100, "登录已过期，请重新登录")
		return
	}
	target := c.Request.URL.RequestURI()
	c.Redirect(http.StatusFound, "/admin/login?redirect="+url.QueryEscape(target))
	c.Abort()
}

// SafeRedirect 只允许站内相对路径，防止开放重定向。
func SafeRedirect(raw, fallback string) string {
	if raw == "" {
		return fallback
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.ContainsAny(raw, "\r\n") {
		return fallback
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host != "" || u.Scheme != "" {
		return fallback
	}
	return raw
}

// SaveRedirect 在 session 中记录登录前路径。
func SaveRedirect(c *gin.Context, target string) {
	s := sessions.Default(c)
	s.Set("login_redirect", SafeRedirect(target, "/admin"))
	_ = s.Save()
}
