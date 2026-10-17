package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cms/internal/service"
)

// MenuHandler 处理后台菜单。
type MenuHandler struct {
	*Base
	Menus *service.MenuService
	Roles *service.RoleService
}

// NewMenuHandler 创建菜单 handler。
func NewMenuHandler(base *Base, menus *service.MenuService, roles *service.RoleService) *MenuHandler {
	return &MenuHandler{Base: base, Menus: menus, Roles: roles}
}

// Index 渲染菜单树。
func (h *MenuHandler) Index(c *gin.Context) {
	h.render(c, "", nil)
}

// Create 创建菜单。
func (h *MenuHandler) Create(c *gin.Context) {
	in := menuInput(c)
	if _, err := h.Menus.Create(c.Request.Context(), in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "菜单已创建", "/admin/menus")
}

// Update 更新菜单。
func (h *MenuHandler) Update(c *gin.Context) {
	in := menuInput(c)
	if _, err := h.Menus.Update(c.Request.Context(), uint64Param(c, "id"), in); err != nil {
		h.render(c, h.ErrorMessage(err), &in)
		return
	}
	h.Success(c, "菜单已保存", "/admin/menus")
}

// Delete 删除菜单。
func (h *MenuHandler) Delete(c *gin.Context) {
	if err := h.Menus.Delete(c.Request.Context(), uint64Param(c, "id")); err != nil {
		h.Fail(c, err, "/admin/menus")
		return
	}
	h.Success(c, "菜单已删除", "/admin/menus")
}

// Reorder 拖拽排序。
func (h *MenuHandler) Reorder(c *gin.Context) {
	if err := h.Menus.Reorder(c.Request.Context(), parseIDs(c.PostFormArray("ids"))); err != nil {
		h.Fail(c, err, "/admin/menus")
		return
	}
	h.JSON(c, gin.H{"ok": true})
}

func (h *MenuHandler) render(c *gin.Context, message string, form *service.MenuInput) {
	tree, err := h.Menus.Tree(c.Request.Context())
	if err != nil {
		h.Fail(c, err, "/admin/menus")
		return
	}
	options, _ := h.Menus.Options(c.Request.Context())
	groups, _ := h.Roles.PermissionGroups(c.Request.Context())
	data := h.Data(c, "菜单管理", "系统管理", "menu")
	data["Tree"] = tree
	data["Options"] = options
	data["Groups"] = groups
	data["Error"] = message
	if form != nil {
		data["Form"] = form
	}
	c.HTML(http.StatusOK, "admin/menu/index.html", data)
}

func menuInput(c *gin.Context) service.MenuInput {
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	return service.MenuInput{
		ParentID:       uint64Param(c, "parent_id"),
		Name:           c.PostForm("name"),
		Icon:           c.PostForm("icon"),
		Path:           c.PostForm("path"),
		PermissionCode: c.PostForm("permission_code"),
		Sort:           sort,
		IsVisible:      c.PostForm("is_visible") == "on",
		IsExternal:     c.PostForm("is_external") == "on",
	}
}
