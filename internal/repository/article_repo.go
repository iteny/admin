// Package repository 封装数据访问；handler 不应直接使用 gorm.DB。
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
)

// ArticleFilter 是文章列表筛选条件。
type ArticleFilter struct {
	pagination.Query
	Status      string
	CategoryID  uint64
	CategoryIDs []uint64
	TagID       uint64
	AuthorID    uint64
	IsTop       *bool
	IsRecommend *bool
}

// ArticleRepo 是文章数据访问。
type ArticleRepo struct{ db *gorm.DB }

// NewArticleRepo 创建文章仓库。
func NewArticleRepo(db *gorm.DB) *ArticleRepo { return &ArticleRepo{db: db} }

// List 返回文章列表和总数。
func (r *ArticleRepo) List(ctx context.Context, f ArticleFilter) ([]model.Article, int64, error) {
	f.Normalize()
	q := r.db.WithContext(ctx).Model(&model.Article{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if len(f.CategoryIDs) > 0 {
		q = q.Where("category_id IN ?", f.CategoryIDs)
	}
	if f.AuthorID > 0 {
		q = q.Where("author_id = ?", f.AuthorID)
	}
	if f.TagID > 0 {
		sub := r.db.Table(r.db.NamingStrategy.JoinTableName("article_tags")).Select("article_id").Where("tag_id = ?", f.TagID)
		q = q.Where("id IN (?)", sub)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("title LIKE ? OR summary LIKE ?", like, like)
	}
	if f.IsTop != nil {
		q = q.Where("is_top = ?", *f.IsTop)
	}
	if f.IsRecommend != nil {
		q = q.Where("is_recommend = ?", *f.IsRecommend)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "COALESCE(published_at, updated_at) DESC, id DESC"
	switch f.Sort {
	case "oldest":
		order = "COALESCE(published_at, created_at) ASC, id ASC"
	case "views":
		order = "view_count DESC, id DESC"
	case "updated":
		order = "updated_at DESC, id DESC"
	}
	var list []model.Article
	err := q.Preload("Category").Preload("Author").Preload("Tags").
		Order(order).Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

// GetByID 按 ID 查询文章。
func (r *ArticleRepo) GetByID(ctx context.Context, id uint64) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").Preload("Tags").First(&article, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &article, err
}

// GetBySlug 按 slug 查询已发布文章。
func (r *ArticleRepo) GetBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").Preload("Tags").
		Where("slug = ? AND status = ?", slug, model.ArticleStatusPublished).First(&article).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &article, err
}

// SlugExists 判断 slug 是否已存在。
func (r *ArticleRepo) SlugExists(ctx context.Context, slug string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Article{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	return n > 0, q.Count(&n).Error
}

// Create 创建文章。
func (r *ArticleRepo) Create(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

// Update 更新文章。
func (r *ArticleRepo) Update(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Save(article).Error
}

// UpdateFields 更新指定字段。
func (r *ArticleRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 软删除文章，并清理标签关联。
func (r *ArticleRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		article := &model.Article{BaseModel: model.BaseModel{ID: id}}
		if err := tx.Model(article).Association("Tags").Clear(); err != nil {
			return err
		}
		return tx.Delete(&model.Article{}, id).Error
	})
}

// DeleteMany 批量软删除，并清理标签关联。
func (r *ArticleRepo) DeleteMany(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			article := &model.Article{BaseModel: model.BaseModel{ID: id}}
			if err := tx.Model(article).Association("Tags").Clear(); err != nil {
				return err
			}
		}
		return tx.Delete(&model.Article{}, ids).Error
	})
}

// IncrementView 原子增加浏览量。
func (r *ArticleRepo) IncrementView(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// Related 返回同分类的相关文章。
func (r *ArticleRepo) Related(ctx context.Context, categoryID, excludeID uint64, limit int) ([]model.Article, error) {
	var list []model.Article
	err := r.db.WithContext(ctx).Where("category_id = ? AND id <> ? AND status = ?", categoryID, excludeID, model.ArticleStatusPublished).
		Order("COALESCE(published_at, created_at) DESC").Limit(limit).Find(&list).Error
	return list, err
}

// PrevNext 返回上一篇和下一篇。
func (r *ArticleRepo) PrevNext(ctx context.Context, article *model.Article) (*model.Article, *model.Article, error) {
	var prev, next model.Article
	published := article.PublishedAt
	if published == nil {
		published = &article.CreatedAt
	}
	errPrev := r.db.WithContext(ctx).Where("status = ? AND COALESCE(published_at, created_at) < ?", model.ArticleStatusPublished, *published).
		Order("COALESCE(published_at, created_at) DESC").First(&prev).Error
	if errors.Is(errPrev, gorm.ErrRecordNotFound) {
		errPrev = nil
	}
	errNext := r.db.WithContext(ctx).Where("status = ? AND COALESCE(published_at, created_at) > ?", model.ArticleStatusPublished, *published).
		Order("COALESCE(published_at, created_at) ASC").First(&next).Error
	if errors.Is(errNext, gorm.ErrRecordNotFound) {
		errNext = nil
	}
	if errPrev != nil {
		return nil, nil, errPrev
	}
	if errNext != nil {
		return nil, nil, errNext
	}
	return &prev, &next, nil
}

// Count 返回文章总数，可按状态过滤。
func (r *ArticleRepo) Count(ctx context.Context, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Article{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

// CountPublishedSince 返回某时间之后发布的文章数。
func (r *ArticleRepo) CountPublishedSince(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("status = ? AND published_at >= ?", model.ArticleStatusPublished, since).Count(&n).Error
	return n, err
}

// Latest 返回最新文章。
func (r *ArticleRepo) Latest(ctx context.Context, limit int) ([]model.Article, error) {
	var list []model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").
		Where("status = ?", model.ArticleStatusPublished).
		Order("COALESCE(published_at, created_at) DESC").Limit(limit).Find(&list).Error
	return list, err
}

// Recent 返回最近文章（包含归档文章，后台仪表盘用）。
func (r *ArticleRepo) Recent(ctx context.Context, limit int) ([]model.Article, error) {
	var list []model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").
		Order("updated_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// SetTags 替换文章标签关联。
func (r *ArticleRepo) SetTags(ctx context.Context, article *model.Article, tags []model.Tag) error {
	return r.db.WithContext(ctx).Model(article).Association("Tags").Replace(tags)
}

// PublishedTrend 返回最近 n 天每日发布数（用数据库无关的逐日 Count）。
func (r *ArticleRepo) PublishedTrend(ctx context.Context, days int) ([]map[string]any, error) {
	now := time.Now().UTC()
	out := make([]map[string]any, 0, days)
	for i := days - 1; i >= 0; i-- {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -i)
		end := start.AddDate(0, 0, 1)
		var n int64
		if err := r.db.WithContext(ctx).Model(&model.Article{}).
			Where("status = ? AND published_at >= ? AND published_at < ?", model.ArticleStatusPublished, start, end).
			Count(&n).Error; err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"date": start.Format("01-02"), "count": n})
	}
	return out, nil
}
