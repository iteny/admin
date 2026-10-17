package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
)

// LogFilter 是操作日志筛选。
type LogFilter struct {
	pagination.Query
	UserID uint64
	Module string
	From   *time.Time
	To     *time.Time
}

// LogRepo 是操作日志数据访问。
type LogRepo struct{ db *gorm.DB }

// NewLogRepo 创建日志仓库。
func NewLogRepo(db *gorm.DB) *LogRepo { return &LogRepo{db: db} }

// Create 写入日志。
func (r *LogRepo) Create(ctx context.Context, log *model.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List 返回日志列表。
func (r *LogRepo) List(ctx context.Context, f LogFilter) ([]model.OperationLog, int64, error) {
	f.Normalize()
	q := r.db.WithContext(ctx).Model(&model.OperationLog{})
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("username LIKE ? OR action LIKE ? OR path LIKE ?", like, like, like)
	}
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Module != "" {
		q = q.Where("module = ?", f.Module)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at < ?", *f.To)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.OperationLog
	err := q.Order("id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// Latest 返回最近日志。
func (r *LogRepo) Latest(ctx context.Context, limit int) ([]model.OperationLog, error) {
	var list []model.OperationLog
	return list, r.db.WithContext(ctx).Order("id DESC").Limit(limit).Find(&list).Error
}

// Clear 清空全部日志。
func (r *LogRepo) Clear(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&model.OperationLog{}).Error
}

// DeleteBefore 删除指定时间之前的日志。
func (r *LogRepo) DeleteBefore(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&model.OperationLog{}).Error
}
