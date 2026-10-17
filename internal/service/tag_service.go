package service

import (
	"context"
	"fmt"
	"strings"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/str"
	"cms/internal/repository"
)

// TagInput 是标签表单输入。
type TagInput struct {
	Name string
	Slug string
}

// TagService 负责标签业务。
type TagService struct {
	repo *repository.TagRepo
}

// NewTagService 创建标签服务。
func NewTagService(repo *repository.TagRepo) *TagService { return &TagService{repo: repo} }

// List 返回标签列表。
func (s *TagService) List(ctx context.Context) ([]model.Tag, error) {
	return s.repo.List(ctx)
}

// Get 按 ID 获取标签。
func (s *TagService) Get(ctx context.Context, id uint64) (*model.Tag, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// GetBySlug 按 slug 获取标签。
func (s *TagService) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// Create 创建标签。
func (s *TagService) Create(ctx context.Context, in TagInput) (*model.Tag, error) {
	item, err := s.build(ctx, in, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return item, nil
}

// Update 更新标签。
func (s *TagService) Update(ctx context.Context, id uint64, in TagInput) (*model.Tag, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	item, err := s.build(ctx, in, id)
	if err != nil {
		return nil, err
	}
	item.ID = id
	item.CreatedAt = existing.CreatedAt
	item.UseCount = existing.UseCount
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}
	return item, nil
}

// Delete 删除标签。
func (s *TagService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

// RecountAll 重新统计使用数。
func (s *TagService) RecountAll(ctx context.Context) error {
	return s.repo.RecountAll(ctx)
}

func (s *TagService) build(ctx context.Context, in TagInput, excludeID uint64) (*model.Tag, error) {
	name := SanitizeText(in.Name)
	if name == "" {
		return nil, apperr.NewValidation("name", "标签名称不能为空")
	}
	if len([]rune(name)) > 50 {
		return nil, apperr.NewValidation("name", "标签名称不能超过 50 个字符")
	}
	exists, err := s.repo.NameExists(ctx, name, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.NewValidation("name", "标签名称已存在")
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = str.Slug(name, true)
	} else {
		slug = str.NormalizeSlug(slug)
	}
	for i := 0; i < 100; i++ {
		candidate := slug
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", slug, i+1)
		}
		exists, err := s.repo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return nil, err
		}
		if !exists {
			slug = candidate
			break
		}
	}
	return &model.Tag{Name: name, Slug: slug}, nil
}
