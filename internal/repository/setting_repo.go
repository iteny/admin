package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"cms/internal/model"
)

// SettingRepo 是键值配置数据访问。
type SettingRepo struct{ db *gorm.DB }

// NewSettingRepo 创建设置仓库。
func NewSettingRepo(db *gorm.DB) *SettingRepo { return &SettingRepo{db: db} }

// All 返回全部配置。
func (r *SettingRepo) All(ctx context.Context) ([]model.Setting, error) {
	var list []model.Setting
	return list, r.db.WithContext(ctx).Order("group_name ASC, id ASC").Find(&list).Error
}

// Get 按 key 查询。
func (r *SettingRepo) Get(ctx context.Context, key string) (*model.Setting, error) {
	var item model.Setting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &item, err
}

// Upsert 保存单个配置。
func (r *SettingRepo) Upsert(ctx context.Context, item *model.Setting) error {
	return r.db.WithContext(ctx).Where("key = ?", item.Key).Assign(item).FirstOrCreate(item).Error
}

// UpsertMany 事务保存多个配置。
func (r *SettingRepo) UpsertMany(ctx context.Context, items []model.Setting) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := tx.Where("key = ?", items[i].Key).Assign(items[i]).FirstOrCreate(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
