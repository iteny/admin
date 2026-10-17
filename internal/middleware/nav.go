package middleware

import (
	"github.com/gin-gonic/gin"

	"cms/internal/model"
	"cms/internal/service"
)

// MenusKey 是后台菜单树在 context 中的键。
const MenusKey = "admin_menus"

// NavCategoriesKey 是前台分类导航键。
const NavCategoriesKey = "nav_categories"

// NavPagesKey 是前台单页导航键。
const NavPagesKey = "nav_pages"

// AdminMenus 按当前用户权限加载侧边栏菜单。
func AdminMenus(menus *service.MenuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tree, err := menus.ForPermissions(c.Request.Context(), Permissions(c))
		if err != nil {
			tree = nil
		}
		c.Set(MenusKey, tree)
		c.Next()
	}
}

// AdminMenuTree 返回侧边栏菜单树。
func AdminMenuTree(c *gin.Context) []*service.MenuNode {
	if v, ok := c.Get(MenusKey); ok {
		if tree, ok := v.([]*service.MenuNode); ok {
			return tree
		}
	}
	return nil
}

// FrontNav 加载前台分类与单页导航。
func FrontNav(front *service.FrontService) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, pages, err := front.Navigation(c.Request.Context())
		if err == nil {
			c.Set(NavCategoriesKey, categories)
			c.Set(NavPagesKey, pages)
		}
		c.Next()
	}
}

// NavCategories 返回前台分类树。
func NavCategories(c *gin.Context) []*service.CategoryNode {
	if v, ok := c.Get(NavCategoriesKey); ok {
		if list, ok := v.([]*service.CategoryNode); ok {
			return list
		}
	}
	return nil
}

// NavPages 返回前台单页导航。
func NavPages(c *gin.Context) []model.Page {
	if v, ok := c.Get(NavPagesKey); ok {
		if list, ok := v.([]model.Page); ok {
			return list
		}
	}
	return nil
}
