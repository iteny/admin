package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
)

// UserRepo 是用户数据访问。
type UserRepo struct{ db *gorm.DB }

// NewUserRepo 创建用户仓库。
func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

// FindByUsername 按用户名查询，并预加载角色权限。
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, err
}

// FindByID 按 ID 查询，并预加载角色权限。
func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Preload("Role.Permissions").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, err
}

// List 返回用户列表。
func (r *UserRepo) List(ctx context.Context, q pagination.Query, status *int8, roleID uint64) ([]model.User, int64, error) {
	q.Normalize()
	db := r.db.WithContext(ctx).Model(&model.User{})
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if roleID > 0 {
		db = db.Where("role_id = ?", roleID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.User
	err := db.Preload("Role").Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

// Create 创建用户。
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户。
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields 更新指定字段。
func (r *UserRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除用户。
func (r *UserRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// Count 返回用户数。
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	return n, r.db.WithContext(ctx).Model(&model.User{}).Count(&n).Error
}

// UsernameExists 判断用户名是否存在。
func (r *UserRepo) UsernameExists(ctx context.Context, username string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// EmailExists 判断邮箱是否存在。
func (r *UserRepo) EmailExists(ctx context.Context, email string, excludeID uint64) (bool, error) {
	if email == "" {
		return false, nil
	}
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// RoleRepo 是角色数据访问。
type RoleRepo struct{ db *gorm.DB }

// NewRoleRepo 创建角色仓库。
func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }

// List 返回角色列表。
func (r *RoleRepo) List(ctx context.Context) ([]model.Role, error) {
	var list []model.Role
	return list, r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
}

// GetByID 按 ID 查询并预加载权限。
func (r *RoleRepo) GetByID(ctx context.Context, id uint64) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &role, err
}

// GetBySlug 按 slug 查询。
func (r *RoleRepo) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Where("slug = ?", slug).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &role, err
}

// Create 创建角色。
func (r *RoleRepo) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update 更新角色。
func (r *RoleRepo) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete 删除角色。
func (r *RoleRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

// CountUsers 统计角色下的用户数。
func (r *RoleRepo) CountUsers(ctx context.Context, id uint64) (int64, error) {
	var n int64
	return n, r.db.WithContext(ctx).Model(&model.User{}).Where("role_id = ?", id).Count(&n).Error
}

// AssignPermissions 替换角色权限。
func (r *RoleRepo) AssignPermissions(ctx context.Context, role *model.Role, perms []model.Permission) error {
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(perms)
}

// SlugExists 判断 slug 是否存在。
func (r *RoleRepo) SlugExists(ctx context.Context, slug string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Role{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// PermissionRepo 是权限点数据访问。
type PermissionRepo struct{ db *gorm.DB }

// NewPermissionRepo 创建权限仓库。
func NewPermissionRepo(db *gorm.DB) *PermissionRepo { return &PermissionRepo{db: db} }

// List 返回全部权限。
func (r *PermissionRepo) List(ctx context.Context) ([]model.Permission, error) {
	var list []model.Permission
	return list, r.db.WithContext(ctx).Order("group_name ASC, sort ASC, id ASC").Find(&list).Error
}

// FindByCodes 按 Code 集合查询。
func (r *PermissionRepo) FindByCodes(ctx context.Context, codes []string) ([]model.Permission, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var list []model.Permission
	return list, r.db.WithContext(ctx).Where("code IN ?", codes).Find(&list).Error
}

// FindByIDs 按 ID 集合查询。
func (r *PermissionRepo) FindByIDs(ctx context.Context, ids []uint64) ([]model.Permission, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.Permission
	return list, r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
}
