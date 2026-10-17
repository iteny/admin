package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
)

// CategoryRepo 是分类数据访问。
type CategoryRepo struct{ db *gorm.DB }

// NewCategoryRepo 创建分类仓库。
func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

// List 返回分类列表。
func (r *CategoryRepo) List(ctx context.Context, onlyEnabled bool) ([]model.Category, error) {
	q := r.db.WithContext(ctx).Order("sort ASC, id ASC")
	if onlyEnabled {
		q = q.Where("status = ?", model.CategoryStatusEnabled)
	}
	var list []model.Category
	return list, q.Find(&list).Error
}

// GetByID 按 ID 查询。
func (r *CategoryRepo) GetByID(ctx context.Context, id uint64) (*model.Category, error) {
	var item model.Category
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// GetBySlug 按 slug 查询启用分类。
func (r *CategoryRepo) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var item model.Category
	err := r.db.WithContext(ctx).Where("slug = ? AND status = ?", slug, model.CategoryStatusEnabled).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Create 创建分类。
func (r *CategoryRepo) Create(ctx context.Context, item *model.Category) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update 更新分类。
func (r *CategoryRepo) Update(ctx context.Context, item *model.Category) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete 删除分类。
func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

// CountChildren 统计直接子分类数。
func (r *CategoryRepo) CountChildren(ctx context.Context, id uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Category{}).Where("parent_id = ?", id).Count(&n).Error
	return n, err
}

// CountArticles 统计分类下文章数。
func (r *CategoryRepo) CountArticles(ctx context.Context, id uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Article{}).Where("category_id = ?", id).Count(&n).Error
	return n, err
}

// ArticleCounts 返回每个分类的文章数。
func (r *CategoryRepo) ArticleCounts(ctx context.Context) (map[uint64]int64, error) {
	type row struct {
		CategoryID uint64
		Count      int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&model.Article{}).
		Select("category_id, COUNT(*) AS count").Group("category_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint64]int64, len(rows))
	for _, item := range rows {
		out[item.CategoryID] = item.Count
	}
	return out, nil
}

// Reorder 按给定 ID 顺序更新 sort。
func (r *CategoryRepo) Reorder(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Category{}).Where("id = ?", id).Update("sort", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SlugExists 判断 slug 是否已存在。
func (r *CategoryRepo) SlugExists(ctx context.Context, slug string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Category{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}
