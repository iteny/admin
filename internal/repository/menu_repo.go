package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
)

// MenuRepo 是后台菜单数据访问。
type MenuRepo struct{ db *gorm.DB }

// NewMenuRepo 创建菜单仓库。
func NewMenuRepo(db *gorm.DB) *MenuRepo { return &MenuRepo{db: db} }

// List 返回全部菜单。
func (r *MenuRepo) List(ctx context.Context) ([]model.Menu, error) {
	var list []model.Menu
	return list, r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
}

// GetByID 按 ID 查询。
func (r *MenuRepo) GetByID(ctx context.Context, id uint64) (*model.Menu, error) {
	var item model.Menu
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Create 创建菜单。
func (r *MenuRepo) Create(ctx context.Context, item *model.Menu) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update 更新菜单。
func (r *MenuRepo) Update(ctx context.Context, item *model.Menu) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete 删除菜单。
func (r *MenuRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}

// CountChildren 统计子菜单数。
func (r *MenuRepo) CountChildren(ctx context.Context, id uint64) (int64, error) {
	var n int64
	return n, r.db.WithContext(ctx).Model(&model.Menu{}).Where("parent_id = ?", id).Count(&n).Error
}

// Reorder 按给定顺序更新 sort。
func (r *MenuRepo) Reorder(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Menu{}).Where("id = ?", id).Update("sort", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MediaRepo 是媒体库数据访问。
type MediaRepo struct{ db *gorm.DB }

// NewMediaRepo 创建媒体仓库。
func NewMediaRepo(db *gorm.DB) *MediaRepo { return &MediaRepo{db: db} }

// List 返回媒体列表。
func (r *MediaRepo) List(ctx context.Context, page, pageSize int, keyword, mimePrefix string) ([]model.Media, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 24
	}
	q := r.db.WithContext(ctx).Model(&model.Media{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("original_name LIKE ? OR stored_name LIKE ?", like, like)
	}
	if mimePrefix != "" {
		q = q.Where("mime_type LIKE ?", mimePrefix+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Media
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetByID 按 ID 查询。
func (r *MediaRepo) GetByID(ctx context.Context, id uint64) (*model.Media, error) {
	var item model.Media
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// FindByHash 按内容摘要查询，用于去重。
func (r *MediaRepo) FindByHash(ctx context.Context, hash string) (*model.Media, error) {
	var item model.Media
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Create 创建媒体记录。
func (r *MediaRepo) Create(ctx context.Context, item *model.Media) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Delete 删除媒体记录。
func (r *MediaRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Media{}, id).Error
}

// Count 返回媒体总数。
func (r *MediaRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	return n, r.db.WithContext(ctx).Model(&model.Media{}).Count(&n).Error
}
