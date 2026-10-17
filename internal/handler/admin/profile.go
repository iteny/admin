package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/pkg/apperr"
	"cms/internal/service"
)

// ProfileHandler 处理个人中心。
type ProfileHandler struct {
	*Base
	Users *service.UserService
	Auth  *service.AuthService
}

// NewProfileHandler 创建个人中心 handler。
func NewProfileHandler(base *Base, users *service.UserService, auth *service.AuthService) *ProfileHandler {
	return &ProfileHandler{Base: base, Users: users, Auth: auth}
}

// Index 渲染个人资料页。
func (h *ProfileHandler) Index(c *gin.Context) {
	user := middleware.CurrentUser(c)
	data := h.Data(c, "个人中心", "账号", "profile")
	data["Profile"] = user
	c.HTML(http.StatusOK, "admin/profile/index.html", data)
}

// Update 更新个人资料。
func (h *ProfileHandler) Update(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		h.Fail(c, apperr.ErrUnauthorized, "/admin/profile")
		return
	}
	if err := h.Users.UpdateProfile(c.Request.Context(), user.ID, c.PostForm("nickname"), c.PostForm("email"), c.PostForm("avatar")); err != nil {
		h.Fail(c, err, "/admin/profile")
		return
	}
	h.Success(c, "个人资料已更新", "/admin/profile")
}

// Password 修改密码。
func (h *ProfileHandler) Password(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		h.Fail(c, apperr.ErrUnauthorized, "/admin/profile")
		return
	}
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")
	confirm := c.PostForm("confirm_password")
	if len(newPassword) < 8 || newPassword != confirm {
		h.Fail(c, apperr.NewValidation("password", "新密码至少 8 位，且两次输入必须一致"), "/admin/profile")
		return
	}
	if err := h.Auth.ChangePassword(c.Request.Context(), user.ID, oldPassword, newPassword); err != nil {
		h.Fail(c, err, "/admin/profile")
		return
	}
	_ = middleware.SessionClear(c)
	c.Redirect(http.StatusFound, "/admin/login")
}
