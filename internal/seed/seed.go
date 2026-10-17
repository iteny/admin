// Package seed 初始化内置角色、权限、管理员和默认菜单。
package seed

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/pkg/hash"
)

// Run 执行幂等种子初始化。
func Run(db *gorm.DB, logger *slog.Logger) error {
	perms, err := ensurePermissions(db)
	if err != nil {
		return err
	}
	roles, err := ensureRoles(db, perms)
	if err != nil {
		return err
	}
	if err := ensureAdmin(db, roles["super-admin"]); err != nil {
		return err
	}
	if err := ensureMenus(db); err != nil {
		return err
	}
	if err := ensureSettings(db); err != nil {
		return err
	}
	logger.Info("seed data ready", "permissions", len(perms), "roles", len(roles))
	return nil
}

var permissionSeed = []model.Permission{
	{Name: "查看仪表盘", Code: "dashboard:view", GroupName: "仪表盘", Sort: 1},
	{Name: "查看文章", Code: "article:view", GroupName: "内容管理", Sort: 10},
	{Name: "创建文章", Code: "article:create", GroupName: "内容管理", Sort: 11},
	{Name: "编辑文章", Code: "article:update", GroupName: "内容管理", Sort: 12},
	{Name: "删除文章", Code: "article:delete", GroupName: "内容管理", Sort: 13},
	{Name: "发布文章", Code: "article:publish", GroupName: "内容管理", Sort: 14},
	{Name: "查看分类", Code: "category:view", GroupName: "内容管理", Sort: 20},
	{Name: "创建分类", Code: "category:create", GroupName: "内容管理", Sort: 21},
	{Name: "编辑分类", Code: "category:update", GroupName: "内容管理", Sort: 22},
	{Name: "删除分类", Code: "category:delete", GroupName: "内容管理", Sort: 23},
	{Name: "查看标签", Code: "tag:view", GroupName: "内容管理", Sort: 30},
	{Name: "创建标签", Code: "tag:create", GroupName: "内容管理", Sort: 31},
	{Name: "编辑标签", Code: "tag:update", GroupName: "内容管理", Sort: 32},
	{Name: "删除标签", Code: "tag:delete", GroupName: "内容管理", Sort: 33},
	{Name: "查看单页", Code: "page:view", GroupName: "内容管理", Sort: 40},
	{Name: "创建单页", Code: "page:create", GroupName: "内容管理", Sort: 41},
	{Name: "编辑单页", Code: "page:update", GroupName: "内容管理", Sort: 42},
	{Name: "删除单页", Code: "page:delete", GroupName: "内容管理", Sort: 43},
	{Name: "查看媒体", Code: "media:view", GroupName: "内容管理", Sort: 50},
	{Name: "上传媒体", Code: "media:upload", GroupName: "内容管理", Sort: 51},
	{Name: "删除媒体", Code: "media:delete", GroupName: "内容管理", Sort: 52},
	{Name: "查看用户", Code: "user:view", GroupName: "系统管理", Sort: 60},
	{Name: "创建用户", Code: "user:create", GroupName: "系统管理", Sort: 61},
	{Name: "编辑用户", Code: "user:update", GroupName: "系统管理", Sort: 62},
	{Name: "删除用户", Code: "user:delete", GroupName: "系统管理", Sort: 63},
	{Name: "重置用户密码", Code: "user:reset_password", GroupName: "系统管理", Sort: 64},
	{Name: "查看角色", Code: "role:view", GroupName: "系统管理", Sort: 70},
	{Name: "创建角色", Code: "role:create", GroupName: "系统管理", Sort: 71},
	{Name: "编辑角色", Code: "role:update", GroupName: "系统管理", Sort: 72},
	{Name: "删除角色", Code: "role:delete", GroupName: "系统管理", Sort: 73},
	{Name: "分配权限", Code: "role:assign", GroupName: "系统管理", Sort: 74},
	{Name: "查看菜单", Code: "menu:view", GroupName: "系统管理", Sort: 80},
	{Name: "创建菜单", Code: "menu:create", GroupName: "系统管理", Sort: 81},
	{Name: "编辑菜单", Code: "menu:update", GroupName: "系统管理", Sort: 82},
	{Name: "删除菜单", Code: "menu:delete", GroupName: "系统管理", Sort: 83},
	{Name: "查看设置", Code: "setting:view", GroupName: "系统管理", Sort: 90},
	{Name: "修改设置", Code: "setting:update", GroupName: "系统管理", Sort: 91},
	{Name: "查看日志", Code: "log:view", GroupName: "系统管理", Sort: 100},
	{Name: "清空日志", Code: "log:delete", GroupName: "系统管理", Sort: 101},
}

func ensurePermissions(db *gorm.DB) (map[string]model.Permission, error) {
	out := make(map[string]model.Permission, len(permissionSeed))
	for _, item := range permissionSeed {
		p := item
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&p).Error; err != nil {
			return nil, fmt.Errorf("seed permission %s: %w", p.Code, err)
		}
		out[p.Code] = p
	}
	return out, nil
}

func ensureRoles(db *gorm.DB, perms map[string]model.Permission) (map[string]*model.Role, error) {
	defs := []struct {
		Name, Slug, Description string
		Codes                   []string
	}{
		{Name: "超级管理员", Slug: "super-admin", Description: "拥有全部权限", Codes: nil},
		{Name: "编辑", Slug: "editor", Description: "负责内容创作与发布", Codes: contentCodes()},
		{Name: "访客", Slug: "viewer", Description: "只读访问", Codes: viewCodes()},
	}
	out := make(map[string]*model.Role, len(defs))
	for _, def := range defs {
		var role model.Role
		err := db.Where("slug = ?", def.Slug).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = model.Role{Name: def.Name, Slug: def.Slug, Description: def.Description, IsSystem: true}
			if err := db.Create(&role).Error; err != nil {
				return nil, fmt.Errorf("seed role %s: %w", def.Slug, err)
			}
			var selected []model.Permission
			if def.Slug == "super-admin" {
				for _, p := range perms {
					selected = append(selected, p)
				}
			} else {
				for _, code := range def.Codes {
					if p, ok := perms[code]; ok {
						selected = append(selected, p)
					}
				}
			}
			if err := db.Model(&role).Association("Permissions").Replace(selected); err != nil {
				return nil, fmt.Errorf("assign role %s permissions: %w", def.Slug, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("load role %s: %w", def.Slug, err)
		} else if def.Slug == "super-admin" {
			// 超级管理员始终拥有全部权限；其他内置角色的权限允许管理员调整。
			selected := make([]model.Permission, 0, len(perms))
			for _, p := range perms {
				selected = append(selected, p)
			}
			if err := db.Model(&role).Association("Permissions").Replace(selected); err != nil {
				return nil, fmt.Errorf("assign role %s permissions: %w", def.Slug, err)
			}
		}
		out[def.Slug] = &role
	}
	return out, nil
}

func contentCodes() []string {
	return []string{
		"dashboard:view",
		"article:view", "article:create", "article:update", "article:delete", "article:publish",
		"category:view", "category:create", "category:update", "category:delete",
		"tag:view", "tag:create", "tag:update", "tag:delete",
		"page:view", "page:create", "page:update", "page:delete",
		"media:view", "media:upload", "media:delete",
	}
}

func viewCodes() []string {
	return []string{
		"dashboard:view", "article:view", "category:view", "tag:view",
		"page:view", "media:view", "user:view", "role:view", "menu:view",
		"setting:view", "log:view",
	}
}

func ensureAdmin(db *gorm.DB, super *model.Role) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	password, err := hash.Password("Admin@123456")
	if err != nil {
		return err
	}
	admin := model.User{
		Username: "admin",
		Password: password,
		Nickname: "超级管理员",
		Email:    "admin@example.com",
		Status:   model.UserStatusEnabled,
		RoleID:   super.ID,
		Remark:   "系统初始化账号，请尽快修改密码",
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

type menuSeed struct {
	Name, Icon, Path, Permission string
	Sort                         int
	Children                     []menuSeed
}

func ensureMenus(db *gorm.DB) error {
	menus := []menuSeed{
		{Name: "仪表盘", Icon: "dashboard", Path: "/admin", Permission: "dashboard:view", Sort: 1},
		{Name: "内容管理", Icon: "file-text", Sort: 10, Children: []menuSeed{
			{Name: "文章", Icon: "article", Path: "/admin/articles", Permission: "article:view", Sort: 11},
			{Name: "分类", Icon: "category", Path: "/admin/categories", Permission: "category:view", Sort: 12},
			{Name: "标签", Icon: "tag", Path: "/admin/tags", Permission: "tag:view", Sort: 13},
			{Name: "单页", Icon: "file", Path: "/admin/pages", Permission: "page:view", Sort: 14},
			{Name: "媒体库", Icon: "photo", Path: "/admin/media", Permission: "media:view", Sort: 15},
		}},
		{Name: "系统管理", Icon: "settings", Sort: 90, Children: []menuSeed{
			{Name: "用户", Icon: "users", Path: "/admin/users", Permission: "user:view", Sort: 91},
			{Name: "角色", Icon: "shield-lock", Path: "/admin/roles", Permission: "role:view", Sort: 92},
			{Name: "菜单", Icon: "menu-2", Path: "/admin/menus", Permission: "menu:view", Sort: 93},
			{Name: "系统设置", Icon: "adjustments", Path: "/admin/settings", Permission: "setting:view", Sort: 94},
			{Name: "操作日志", Icon: "history", Path: "/admin/logs", Permission: "log:view", Sort: 95},
		}},
	}
	var create func(parentID uint64, items []menuSeed) error
	create = func(parentID uint64, items []menuSeed) error {
		for _, item := range items {
			var menu model.Menu
			seedMenu := model.Menu{
				ParentID: parentID, Name: item.Name, Icon: item.Icon, Path: item.Path,
				PermissionCode: item.Permission, Sort: item.Sort, IsVisible: true,
			}
			if err := db.Where("name = ? AND parent_id = ?", item.Name, parentID).FirstOrCreate(&menu, seedMenu).Error; err != nil {
				return fmt.Errorf("seed menu %s: %w", item.Name, err)
			}
			if len(item.Children) > 0 {
				if err := create(menu.ID, item.Children); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return create(0, menus)
}

func ensureSettings(db *gorm.DB) error {
	defaults := map[string]string{
		"site_name":          "Tabler CMS",
		"site_logo":          "",
		"site_favicon":       "",
		"site_icp":           "",
		"site_footer":        "© 2026 Tabler CMS",
		"site_description":   "一个使用 Gin、GORM 与 Tabler 构建的内容管理系统",
		"seo_title":          "Tabler CMS",
		"seo_keywords":       "CMS,Gin,GORM,Tabler",
		"seo_description":    "一个使用 Gin、GORM 与 Tabler 构建的内容管理系统",
		"upload_max_size":    "10485760",
		"upload_allowed_ext": "jpg,jpeg,png,gif,webp,svg",
		"mail_host":          "",
		"mail_port":          "587",
		"mail_username":      "",
		"mail_password":      "",
		"mail_from":          "",
	}
	for key, value := range defaults {
		group := "site"
		switch {
		case strings.HasPrefix(key, "seo_"):
			group = "seo"
		case strings.HasPrefix(key, "upload"):
			group = "upload"
		case strings.HasPrefix(key, "mail"):
			group = "mail"
		}
		var item model.Setting
		err := db.Where("key = ?", key).First(&item).Error
		if err == nil {
			// 已存在时不覆盖用户修改，只补齐分组/类型。
			_ = db.Model(&item).Updates(map[string]any{"group_name": group, "type": "string"}).Error
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("seed setting %s: %w", key, err)
		}
		if err := db.Create(&model.Setting{Key: key, Value: value, GroupName: group, Type: "string"}).Error; err != nil {
			return fmt.Errorf("seed setting %s: %w", key, err)
		}
	}
	return nil
}
