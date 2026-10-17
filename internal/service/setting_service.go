// Package service 承载业务规则和事务边界。
package service

import (
	"context"
	"fmt"
	"sync"

	"cms/internal/model"
	"cms/internal/repository"
)

// SettingService 缓存键值配置，保存后刷新。
type SettingService struct {
	repo  *repository.SettingRepo
	mu    sync.RWMutex
	cache map[string]string
}

// NewSettingService 创建设置服务。
func NewSettingService(repo *repository.SettingRepo) *SettingService {
	return &SettingService{repo: repo, cache: map[string]string{}}
}

// Load 从数据库加载全部配置到内存。
func (s *SettingService) Load(ctx context.Context) error {
	list, err := s.repo.All(ctx)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	cache := make(map[string]string, len(list))
	for _, item := range list {
		cache[item.Key] = item.Value
	}
	s.mu.Lock()
	s.cache = cache
	s.mu.Unlock()
	return nil
}

// Get 返回配置值，不存在时返回默认值。
func (s *SettingService) Get(key, def string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.cache[key]; ok && v != "" {
		return v
	}
	return def
}

// All 返回配置快照。
func (s *SettingService) All() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.cache))
	for k, v := range s.cache {
		out[k] = v
	}
	return out
}

// List 返回数据库中的配置项（设置页用）。
func (s *SettingService) List(ctx context.Context) ([]model.Setting, error) {
	return s.repo.All(ctx)
}

// Save 保存一组配置并刷新缓存。
func (s *SettingService) Save(ctx context.Context, values map[string]string, group string) error {
	items := make([]model.Setting, 0, len(values))
	for key, value := range values {
		items = append(items, model.Setting{Key: key, Value: value, GroupName: group, Type: "string"})
	}
	if err := s.repo.UpsertMany(ctx, items); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	return s.Load(ctx)
}
