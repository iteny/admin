package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cms/internal/model"
	"cms/internal/pkg/hash"
	"cms/internal/repository"
)

var (
	// ErrInvalidCredentials 表示用户名或密码错误（不区分具体原因）。
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	// ErrUserDisabled 表示账号被禁用。
	ErrUserDisabled = errors.New("账号已被禁用")
	// ErrTooManyAttempts 表示登录失败次数过多。
	ErrTooManyAttempts = errors.New("登录失败次数过多，请稍后再试")
)

// AuthService 负责登录、权限加载和密码修改。
type AuthService struct {
	users    *repository.UserRepo
	limiter  *loginLimiter
	attempts int
	window   time.Duration
}

// NewAuthService 创建认证服务。
func NewAuthService(users *repository.UserRepo) *AuthService {
	return &AuthService{
		users:    users,
		limiter:  newLoginLimiter(),
		attempts: 5,
		window:   15 * time.Minute,
	}
}

// Authenticate 校验用户名密码，并返回带角色权限的用户。
func (s *AuthService) Authenticate(ctx context.Context, username, password, ip string) (*model.User, error) {
	key := username + "|" + ip
	if ok, wait := s.limiter.Allowed(key); !ok {
		return nil, fmt.Errorf("%w（%d 分钟后重试）", ErrTooManyAttempts, int(wait.Minutes())+1)
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		s.limiter.Fail(key, s.attempts, s.window)
		return nil, ErrInvalidCredentials
	}
	if !hash.CheckPassword(user.Password, password) {
		s.limiter.Fail(key, s.attempts, s.window)
		return nil, ErrInvalidCredentials
	}
	if user.Status != model.UserStatusEnabled {
		return nil, ErrUserDisabled
	}
	s.limiter.Reset(key)
	now := time.Now().UTC()
	_ = s.users.UpdateFields(ctx, user.ID, map[string]any{"last_login_at": now, "last_login_ip": ip})
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	return user, nil
}

// CurrentUser 按 ID 加载用户，并预加载角色权限。
func (s *AuthService) CurrentUser(ctx context.Context, id uint64) (*model.User, error) {
	return s.users.FindByID(ctx, id)
}

// Permissions 返回权限集合；超级管理员返回通配权限。
func (s *AuthService) Permissions(user *model.User) map[string]bool {
	perms := map[string]bool{}
	if user == nil {
		return perms
	}
	if user.Role.Slug == "super-admin" {
		perms["*"] = true
		return perms
	}
	for _, p := range user.Role.Permissions {
		perms[p.Code] = true
	}
	return perms
}

// ChangePassword 校验原密码并更新为新密码。
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("新密码至少 8 位")
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("load user: %w", err)
	}
	if !hash.CheckPassword(user.Password, oldPassword) {
		return errors.New("原密码不正确")
	}
	hashed, err := hash.Password(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdateFields(ctx, userID, map[string]any{"password": hashed})
}

// ResetPassword 由管理员重置用户密码。
func (s *AuthService) ResetPassword(ctx context.Context, userID uint64, password string) error {
	if len(password) < 8 {
		return errors.New("密码至少 8 位")
	}
	hashed, err := hash.Password(password)
	if err != nil {
		return err
	}
	return s.users.UpdateFields(ctx, userID, map[string]any{"password": hashed})
}

// loginLimiter 是按用户名+IP 的内存失败计数器。
type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*attemptEntry
}

type attemptEntry struct {
	count int
	until time.Time
}

func newLoginLimiter() *loginLimiter {
	l := &loginLimiter{entries: make(map[string]*attemptEntry)}
	go l.cleanup()
	return l
}

func (l *loginLimiter) Allowed(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.entries[key]
	if !ok {
		return true, 0
	}
	if time.Now().After(entry.until) {
		delete(l.entries, key)
		return true, 0
	}
	if entry.count >= 5 {
		return false, time.Until(entry.until)
	}
	return true, 0
}

func (l *loginLimiter) Fail(key string, attempts int, window time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.entries[key]
	if !ok || time.Now().After(entry.until) {
		entry = &attemptEntry{}
		l.entries[key] = entry
	}
	entry.count++
	if entry.count >= attempts {
		entry.until = time.Now().Add(window)
	}
}

func (l *loginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func (l *loginLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, entry := range l.entries {
			if now.After(entry.until) && entry.until != (time.Time{}) {
				delete(l.entries, key)
			}
		}
		l.mu.Unlock()
	}
}
