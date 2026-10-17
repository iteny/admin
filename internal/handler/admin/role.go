package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cms/internal/model"
	"cms/internal/service"
)

// RoleHandler 处理角色与权限。
type RoleHandler struct {
	*Base
	Roles *service.RoleService
}

// NewRoleHandler 创建角色 handler。
func NewRoleHandler(base *Base, roles *service.RoleService) *RoleHandler {
	return &RoleHandler{Base: base, Roles: roles}
}

// Index 渲染角色列表。
func (h *RoleHandler) Index(c *gin.Context) {
	roles, err := h.Roles.List(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin/roles")
		return
	}
	data := h.Data(c, "角色管理", "系统管理", "role")
	data["Roles"] = roles
	c.HTML(http.StatusOK, "admin/role/index.html", data)
}

// CreatePage 渲染新增角色页。
func (h *RoleHandler) CreatePage(c *gin.Context) {
	h.renderForm(c, &model.Role{}, nil, "")
}

// Create 创建角色。
func (h *RoleHandler) Create(c *gin.Context) {
	role, err := h.Roles.Create(c.Request.Context(), c.PostForm("name"), c.PostForm("slug"), c.PostForm("description"), parseIDs(c.PostFormArray("permission_ids")))
	if err != nil {
		selected := make(map[uint64]bool)
		for _, id := range parseIDs(c.PostFormArray("permission_ids")) {
			selected[id] = true
		}
		h.renderForm(c, roleFromForm(c, 0), selected, h.ErrorMessage(err))
		return
	}
	h.Success(c, "角色已创建", "/admin/roles/"+strconv.FormatUint(role.ID, 10)+"/edit")
}

// EditPage 渲染编辑页。
func (h *RoleHandler) EditPage(c *gin.Context) {
	role, err := h.Roles.Get(c.Request.Context(), uint64Param(c, "id"))
	if err != nil {
		h.Fail(c, err, "/admin/roles")
		return
	}
	selected := make(map[uint64]bool, len(role.Permissions))
	for _, p := range role.Permissions {
		selected[p.ID] = true
	}
	h.renderForm(c, role, selected, "")
}

// Update 更新角色。
func (h *RoleHandler) Update(c *gin.Context) {
	id := uint64Param(c, "id")
	role, err := h.Roles.Update(c.Request.Context(), id, c.PostForm("name"), c.PostForm("slug"), c.PostForm("description"), parseIDs(c.PostFormArray("permission_ids")))
	if err != nil {
		selected := make(map[uint64]bool)
		for _, permissionID := range parseIDs(c.PostFormArray("permission_ids")) {
			selected[permissionID] = true
		}
		h.renderForm(c, roleFromForm(c, id), selected, h.ErrorMessage(err))
		return
	}
	h.Success(c, "角色已保存", "/admin/roles/"+strconv.FormatUint(role.ID, 10)+"/edit")
}

// Delete 删除角色。
func (h *RoleHandler) Delete(c *gin.Context) {
	if err := h.Roles.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		h.Fail(c, err, "/admin/roles")
		return
	}
	h.Success(c, "角色已删除", "/admin/roles")
}

func (h *RoleHandler) renderForm(c *gin.Context, role *model.Role, selected map[uint64]bool, message string) {
	groups, _ := h.Roles.PermissionGroups(c.Request.Context())
	title := "新增角色"
	if role != nil && role.ID > 0 {
		title = "编辑角色"
	}
	data := h.Data(c, title, "系统管理", "role")
	data["Role"] = role
	data["Groups"] = groups
	data["Selected"] = selected
	data["Error"] = message
	data["IsEdit"] = role != nil && role.ID > 0
	data["ActionURL"] = "/admin/roles"
	if role != nil && role.ID > 0 {
		data["ActionURL"] = "/admin/roles/" + strconv.FormatUint(role.ID, 10)
	}
	c.HTML(http.StatusOK, "admin/role/form.html", data)
}

func roleFromForm(c *gin.Context, id uint64) *model.Role {
	return &model.Role{
		BaseModel:   model.BaseModel{ID: id},
		Name:        c.PostForm("name"),
		Slug:        c.PostForm("slug"),
		Description: c.PostForm("description"),
	}
}
