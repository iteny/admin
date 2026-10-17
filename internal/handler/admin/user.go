package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/pagination"
	"cms/internal/service"
)

// UserHandler 处理用户管理。
type UserHandler struct {
	*Base
	Users *service.UserService
	Roles *service.RoleService
}

// NewUserHandler 创建用户 handler。
func NewUserHandler(base *Base, users *service.UserService, roles *service.RoleService) *UserHandler {
	return &UserHandler{Base: base, Users: users, Roles: roles}
}

// Index 渲染用户列表。
func (h *UserHandler) Index(c *gin.Context) {
	q := pagination.Query{Page: intParam(c, "page", 1), PageSize: intParam(c, "page_size", 15), Keyword: c.Query("keyword")}
	var status *int8
	if raw := c.Query("status"); raw == "0" || raw == "1" {
		v := int8(raw[0] - '0')
		status = &v
	}
	roleID := uint64Param(c, "role_id")
	list, total, err := h.Users.List(c.Request.Context(), q, status, roleID)
	if err != nil {
		h.Fail(c, err, "/admin/users")
		return
	}
	roles, _ := h.Roles.List(c.Request.Context())
	data := h.Data(c, "用户管理", "系统管理", "user")
	data["Data"] = pagination.New(list, total, q.Page, q.PageSize)
	data["Roles"] = roles
	data["Status"] = c.Query("status")
	data["RoleID"] = roleID
	data["QueryValues"] = c.Request.URL.Query()
	data["BaseURL"] = "/admin/users"
	c.HTML(http.StatusOK, "admin/user/index.html", data)
}

// CreatePage 渲染新增用户页。
func (h *UserHandler) CreatePage(c *gin.Context) {
	h.renderForm(c, &model.User{Status: model.UserStatusEnabled}, "")
}

// Create 创建用户。
func (h *UserHandler) Create(c *gin.Context) {
	in := userInput(c)
	user, err := h.Users.Create(c.Request.Context(), in)
	if err != nil {
		h.renderForm(c, userFromInput(0, in), h.ErrorMessage(err))
		return
	}
	h.Success(c, "用户已创建", "/admin/users/"+strconv.FormatUint(user.ID, 10)+"/edit")
}

// EditPage 渲染编辑页。
func (h *UserHandler) EditPage(c *gin.Context) {
	user, err := h.Users.Get(c.Request.Context(), uint64Param(c, "id"))
	if err != nil {
		h.Fail(c, err, "/admin/users")
		return
	}
	h.renderForm(c, user, "")
}

// Update 更新用户。
func (h *UserHandler) Update(c *gin.Context) {
	id := uint64Param(c, "id")
	in := userInput(c)
	user, err := h.Users.Update(c.Request.Context(), id, in)
	if err != nil {
		h.renderForm(c, userFromInput(id, in), h.ErrorMessage(err))
		return
	}
	h.Success(c, "用户已保存", "/admin/users/"+strconv.FormatUint(user.ID, 10)+"/edit")
}

// Delete 删除用户。
func (h *UserHandler) Delete(c *gin.Context) {
	current := middleware.CurrentUser(c)
	if current == nil {
		h.Fail(c, apperr.ErrUnauthorized, "/admin/users")
		return
	}
	if err := h.Users.Delete(c.Request.Context(), uint64Param(c, "id"), current.ID); err != nil {
		h.Fail(c, err, "/admin/users")
		return
	}
	h.Success(c, "用户已删除", "/admin/users")
}

// ResetPassword 重置密码。
func (h *UserHandler) ResetPassword(c *gin.Context) {
	password := c.PostForm("password")
	if err := h.Users.ResetPassword(c.Request.Context(), uint64Param(c, "id"), password); err != nil {
		h.Fail(c, err, "/admin/users")
		return
	}
	h.Success(c, "密码已重置", "/admin/users")
}

func (h *UserHandler) renderForm(c *gin.Context, user *model.User, message string) {
	roles, _ := h.Roles.List(c.Request.Context())
	title := "新增用户"
	if user != nil && user.ID > 0 {
		title = "编辑用户"
	}
	data := h.Data(c, title, "系统管理", "user")
	data["User"] = user
	data["Roles"] = roles
	data["Error"] = message
	data["IsEdit"] = user != nil && user.ID > 0
	data["ActionURL"] = "/admin/users"
	if user != nil && user.ID > 0 {
		data["ActionURL"] = "/admin/users/" + strconv.FormatUint(user.ID, 10)
	}
	c.HTML(http.StatusOK, "admin/user/form.html", data)
}

func userInput(c *gin.Context) service.UserInput {
	status := int8(1)
	if c.PostForm("status") == "0" {
		status = 0
	}
	return service.UserInput{
		Username: c.PostForm("username"),
		Nickname: c.PostForm("nickname"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
		Avatar:   c.PostForm("avatar"),
		Status:   status,
		RoleID:   uint64Param(c, "role_id"),
		Remark:   c.PostForm("remark"),
	}
}

func userFromInput(id uint64, in service.UserInput) *model.User {
	return &model.User{
		BaseModel: model.BaseModel{ID: id},
		Username:  in.Username, Nickname: in.Nickname, Email: in.Email,
		Avatar: in.Avatar, Status: in.Status, RoleID: in.RoleID, Remark: in.Remark,
	}
}
