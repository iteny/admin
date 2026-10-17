package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/hash"
	"cms/internal/pkg/pagination"
	"cms/internal/repository"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)

// UserInput 是用户表单输入。
type UserInput struct {
	Username string
	Nickname string
	Email    string
	Password string
	Avatar   string
	Status   int8
	RoleID   uint64
	Remark   string
}

// UserService 负责用户管理。
type UserService struct {
	users *repository.UserRepo
	roles *repository.RoleRepo
}

// NewUserService 创建用户服务。
func NewUserService(users *repository.UserRepo, roles *repository.RoleRepo) *UserService {
	return &UserService{users: users, roles: roles}
}

// List 返回用户列表。
func (s *UserService) List(ctx context.Context, q pagination.Query, status *int8, roleID uint64) ([]model.User, int64, error) {
	return s.users.List(ctx, q, status, roleID)
}

// Get 按 ID 获取用户。
func (s *UserService) Get(ctx context.Context, id uint64) (*model.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return user, nil
}

// Create 创建用户。
func (s *UserService) Create(ctx context.Context, in UserInput) (*model.User, error) {
	user, err := s.build(ctx, in, 0)
	if err != nil {
		return nil, err
	}
	if len(user.Password) < 8 {
		return nil, apperr.NewValidation("password", "密码至少 8 位")
	}
	hashed, err := hash.Password(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashed
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Update 更新用户；密码为空时保持不变。
func (s *UserService) Update(ctx context.Context, id uint64, in UserInput) (*model.User, error) {
	existing, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	user, err := s.build(ctx, in, id)
	if err != nil {
		return nil, err
	}
	user.ID = id
	user.CreatedAt = existing.CreatedAt
	user.Password = existing.Password
	user.LastLoginAt = existing.LastLoginAt
	user.LastLoginIP = existing.LastLoginIP
	user.Avatar = strings.TrimSpace(in.Avatar)
	if in.Password != "" {
		if len(in.Password) < 8 {
			return nil, apperr.NewValidation("password", "密码至少 8 位")
		}
		hashed, err := hash.Password(in.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

// Delete 删除用户，禁止删除自己和超级管理员。
func (s *UserService) Delete(ctx context.Context, id, currentUserID uint64) error {
	if id == currentUserID {
		return apperr.NewValidation("id", "不能删除自己")
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if user.Role.Slug == "super-admin" {
		return apperr.NewValidation("id", "不能删除超级管理员")
	}
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// ResetPassword 重置密码。
func (s *UserService) ResetPassword(ctx context.Context, id uint64, password string) error {
	if len(password) < 8 {
		return apperr.NewValidation("password", "密码至少 8 位")
	}
	hashed, err := hash.Password(password)
	if err != nil {
		return err
	}
	return s.users.UpdateFields(ctx, id, map[string]any{"password": hashed})
}

// UpdateProfile 更新个人资料。
func (s *UserService) UpdateProfile(ctx context.Context, id uint64, nickname, email, avatar string) error {
	if email != "" {
		exists, err := s.users.EmailExists(ctx, email, id)
		if err != nil {
			return err
		}
		if exists {
			return apperr.NewValidation("email", "邮箱已被使用")
		}
	}
	return s.users.UpdateFields(ctx, id, map[string]any{
		"nickname": strings.TrimSpace(nickname),
		"email":    strings.TrimSpace(email),
		"avatar":   strings.TrimSpace(avatar),
	})
}

func (s *UserService) build(ctx context.Context, in UserInput, excludeID uint64) (*model.User, error) {
	username := strings.TrimSpace(in.Username)
	if !usernamePattern.MatchString(username) {
		return nil, apperr.NewValidation("username", "用户名只能包含字母、数字、下划线，长度 3-50")
	}
	exists, err := s.users.UsernameExists(ctx, username, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.NewValidation("username", "用户名已存在")
	}
	email := strings.TrimSpace(in.Email)
	if email != "" {
		exists, err := s.users.EmailExists(ctx, email, excludeID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperr.NewValidation("email", "邮箱已被使用")
		}
	}
	if in.RoleID == 0 {
		return nil, apperr.NewValidation("role_id", "请选择角色")
	}
	if _, err := s.roles.GetByID(ctx, in.RoleID); err != nil {
		return nil, apperr.NewValidation("role_id", "角色不存在")
	}
	status := in.Status
	if status != model.UserStatusDisabled {
		status = model.UserStatusEnabled
	}
	return &model.User{
		Username: username,
		Nickname: SanitizeText(in.Nickname),
		Email:    email,
		Password: in.Password,
		Avatar:   strings.TrimSpace(in.Avatar),
		Status:   status,
		RoleID:   in.RoleID,
		Remark:   SanitizeText(in.Remark),
	}, nil
}

// RoleService 负责角色与权限。
type RoleService struct {
	roles       *repository.RoleRepo
	permissions *repository.PermissionRepo
}

// NewRoleService 创建角色服务。
func NewRoleService(roles *repository.RoleRepo, permissions *repository.PermissionRepo) *RoleService {
	return &RoleService{roles: roles, permissions: permissions}
}

// List 返回角色列表。
func (s *RoleService) List(ctx context.Context) ([]model.Role, error) {
	return s.roles.List(ctx)
}

// Get 按 ID 获取角色。
func (s *RoleService) Get(ctx context.Context, id uint64) (*model.Role, error) {
	role, err := s.roles.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return role, nil
}

// AllPermissions 返回全部权限点。
func (s *RoleService) AllPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.permissions.List(ctx)
}

// PermissionGroup 是权限分组。
type PermissionGroup struct {
	Name  string
	Perms []model.Permission
}

// PermissionGroups 返回分组后的权限树。
func (s *RoleService) PermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
	list, err := s.permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	order := make([]string, 0)
	groups := make(map[string][]model.Permission)
	for _, item := range list {
		group := item.GroupName
		if group == "" {
			group = "其他"
		}
		if _, ok := groups[group]; !ok {
			order = append(order, group)
		}
		groups[group] = append(groups[group], item)
	}
	out := make([]PermissionGroup, 0, len(order))
	for _, name := range order {
		out = append(out, PermissionGroup{Name: name, Perms: groups[name]})
	}
	return out, nil
}

// Create 创建角色。
func (s *RoleService) Create(ctx context.Context, name, slug, description string, permissionIDs []uint64) (*model.Role, error) {
	name = SanitizeText(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, apperr.NewValidation("name", "角色名称不能为空")
	}
	if slug == "" {
		return nil, apperr.NewValidation("slug", "角色标识不能为空")
	}
	exists, err := s.roles.SlugExists(ctx, slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.NewValidation("slug", "角色标识已存在")
	}
	role := &model.Role{Name: name, Slug: slug, Description: SanitizeText(description)}
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	if err := s.assign(ctx, role, permissionIDs); err != nil {
		return nil, err
	}
	return s.roles.GetByID(ctx, role.ID)
}

// Update 更新角色。
func (s *RoleService) Update(ctx context.Context, id uint64, name, slug, description string, permissionIDs []uint64) (*model.Role, error) {
	role, err := s.roles.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	name = SanitizeText(name)
	slug = strings.TrimSpace(slug)
	if name == "" || slug == "" {
		return nil, apperr.NewValidation("name", "角色名称和标识不能为空")
	}
	if role.IsSystem {
		slug = role.Slug
	} else {
		exists, err := s.roles.SlugExists(ctx, slug, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperr.NewValidation("slug", "角色标识已存在")
		}
	}
	role.Name = name
	role.Slug = strings.TrimSpace(slug)
	role.Description = SanitizeText(description)
	if err := s.roles.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	if err := s.assign(ctx, role, permissionIDs); err != nil {
		return nil, err
	}
	return s.roles.GetByID(ctx, id)
}

// Delete 删除角色。
func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	role, err := s.roles.GetByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if role.IsSystem {
		return apperr.NewValidation("id", "内置角色不能删除")
	}
	n, err := s.roles.CountUsers(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperr.NewValidation("id", "该角色下还有用户，不能删除")
	}
	if err := s.roles.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *RoleService) assign(ctx context.Context, role *model.Role, ids []uint64) error {
	perms, err := s.permissions.FindByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("load permissions: %w", err)
	}
	if role.Slug == "super-admin" {
		perms, err = s.permissions.List(ctx)
		if err != nil {
			return err
		}
	}
	if err := s.roles.AssignPermissions(ctx, role, perms); err != nil {
		return fmt.Errorf("assign permissions: %w", err)
	}
	return nil
}
