package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
)

// PageRepo 是单页数据访问。
type PageRepo struct{ db *gorm.DB }

// NewPageRepo 创建单页仓库。
func NewPageRepo(db *gorm.DB) *PageRepo { return &PageRepo{db: db} }

// List 返回单页列表。
func (r *PageRepo) List(ctx context.Context, q pagination.Query) ([]model.Page, int64, error) {
	q.Normalize()
	db := r.db.WithContext(ctx).Model(&model.Page{})
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("title LIKE ? OR slug LIKE ?", like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Page
	err := db.Order("sort ASC, id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

// GetByID 按 ID 查询。
func (r *PageRepo) GetByID(ctx context.Context, id uint64) (*model.Page, error) {
	var item model.Page
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// GetBySlug 按 slug 查询已发布单页。
func (r *PageRepo) GetBySlug(ctx context.Context, slug string) (*model.Page, error) {
	var item model.Page
	err := r.db.WithContext(ctx).Where("slug = ? AND status = ?", slug, model.PageStatusPublished).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Create 创建单页。
func (r *PageRepo) Create(ctx context.Context, item *model.Page) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update 更新单页。
func (r *PageRepo) Update(ctx context.Context, item *model.Page) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete 删除单页。
func (r *PageRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Page{}, id).Error
}

// Published 返回全部已发布单页（前台导航用）。
func (r *PageRepo) Published(ctx context.Context) ([]model.Page, error) {
	var list []model.Page
	return list, r.db.WithContext(ctx).Where("status = ?", model.PageStatusPublished).Order("sort ASC, id ASC").Find(&list).Error
}

// SlugExists 判断 slug 是否已存在。
func (r *PageRepo) SlugExists(ctx context.Context, slug string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Page{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}
