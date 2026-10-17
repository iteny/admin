package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
)

// TagRepo 是标签数据访问。
type TagRepo struct{ db *gorm.DB }

// NewTagRepo 创建标签仓库。
func NewTagRepo(db *gorm.DB) *TagRepo { return &TagRepo{db: db} }

// List 返回全部标签。
func (r *TagRepo) List(ctx context.Context) ([]model.Tag, error) {
	var list []model.Tag
	return list, r.db.WithContext(ctx).Order("use_count DESC, id DESC").Find(&list).Error
}

// FindByIDs 按 ID 集合查询。
func (r *TagRepo) FindByIDs(ctx context.Context, ids []uint64) ([]model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.Tag
	return list, r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
}

// FindByNames 按名称集合查询。
func (r *TagRepo) FindByNames(ctx context.Context, names []string) ([]model.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var list []model.Tag
	return list, r.db.WithContext(ctx).Where("name IN ?", names).Find(&list).Error
}

// GetByID 按 ID 查询。
func (r *TagRepo) GetByID(ctx context.Context, id uint64) (*model.Tag, error) {
	var item model.Tag
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// GetBySlug 按 slug 查询。
func (r *TagRepo) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	var item model.Tag
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Create 创建标签。
func (r *TagRepo) Create(ctx context.Context, item *model.Tag) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update 更新标签。
func (r *TagRepo) Update(ctx context.Context, item *model.Tag) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete 删除标签并清理关联。
func (r *TagRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(tx.NamingStrategy.JoinTableName("article_tags")).Where("tag_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Tag{}, id).Error
	})
}

// SlugExists 判断 slug 是否已存在。
func (r *TagRepo) SlugExists(ctx context.Context, slug string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Tag{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// NameExists 判断名称是否已存在。
func (r *TagRepo) NameExists(ctx context.Context, name string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Tag{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// Recount 重新计算标签使用数。
func (r *TagRepo) Recount(ctx context.Context, id uint64) error {
	table := r.db.NamingStrategy.JoinTableName("article_tags")
	var n int64
	if err := r.db.WithContext(ctx).Table(table).Where("tag_id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id).Update("use_count", n).Error
}

// RecountAll 重新计算全部标签使用数。
func (r *TagRepo) RecountAll(ctx context.Context) error {
	var tags []model.Tag
	if err := r.db.WithContext(ctx).Find(&tags).Error; err != nil {
		return err
	}
	for _, tag := range tags {
		if err := r.Recount(ctx, tag.ID); err != nil {
			return err
		}
	}
	return nil
}
