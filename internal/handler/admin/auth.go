package admin

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/pkg/flash"
	"cms/internal/service"
)

// AuthHandler 处理登录与退出。
type AuthHandler struct {
	*Base
	Auth *service.AuthService
}

// NewAuthHandler 创建认证 handler。
func NewAuthHandler(base *Base, auth *service.AuthService) *AuthHandler {
	return &AuthHandler{Base: base, Auth: auth}
}

// LoginPage 渲染登录页。
func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/auth/login.html", gin.H{
		"Title":     "登录",
		"SiteName":  middleware.SiteName(c),
		"CSRFToken": middleware.Token(c),
		"Redirect":  middleware.SafeRedirect(c.Query("redirect"), "/admin"),
		"Username":  "",
		"Error":     "",
	})
}

// Login 校验登录。
func (h *AuthHandler) Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	remember := c.PostForm("remember") == "on"
	redirect := middleware.SafeRedirect(c.PostForm("redirect"), "/admin")

	renderError := func(message string) {
		c.HTML(http.StatusOK, "admin/auth/login.html", gin.H{
			"Title":     "登录",
			"SiteName":  middleware.SiteName(c),
			"CSRFToken": middleware.Token(c),
			"Redirect":  redirect,
			"Username":  username,
			"Error":     message,
		})
	}
	if username == "" || password == "" {
		renderError("请输入用户名和密码")
		return
	}

	user, err := h.Auth.Authenticate(c.Request.Context(), username, password, c.ClientIP())
	if err != nil {
		renderError(h.ErrorMessage(err))
		return
	}

	s := sessions.Default(c)
	s.Clear()
	s.Set("uid", user.ID)
	age := 7 * 24 * 3600
	if remember {
		age = 30 * 24 * 3600
	}
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   age,
		HttpOnly: true,
		Secure:   h.Config.Session.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	if err := s.Save(); err != nil {
		h.Logger.Error("save session failed", "error", err)
		renderError("登录成功，但会话保存失败，请检查 Cookie 设置")
		return
	}
	flash.Success(c, "欢迎回来，"+user.Nickname)
	c.Redirect(http.StatusFound, redirect)
}

// Logout 退出登录。
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := middleware.SessionClear(c); err != nil {
		h.Logger.Warn("clear session failed", "error", err)
	}
	c.Redirect(http.StatusFound, "/admin/login")
}
