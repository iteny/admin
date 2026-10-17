package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"gorm.io/gorm"

	"cms/internal/model"
	"cms/internal/repository"
)

// DashboardStats 是仪表盘统计。
type DashboardStats struct {
	Articles       int64
	Published      int64
	Categories     int64
	Users          int64
	Media          int64
	Views          int64
	TodayPublished int64
	Trend          []map[string]any
	Latest         []model.Article
	Logs           []model.OperationLog
	GoVersion      string
	Uptime         string
	DBVersion      string
}

// DashboardService 汇总仪表盘数据。
type DashboardService struct {
	db         *gorm.DB
	articles   *repository.ArticleRepo
	categories *repository.CategoryRepo
	users      *repository.UserRepo
	media      *repository.MediaRepo
	logs       *repository.LogRepo
	startedAt  time.Time
}

// NewDashboardService 创建仪表盘服务。
func NewDashboardService(db *gorm.DB, articles *repository.ArticleRepo, categories *repository.CategoryRepo, users *repository.UserRepo, media *repository.MediaRepo, logs *repository.LogRepo) *DashboardService {
	return &DashboardService{db: db, articles: articles, categories: categories, users: users, media: media, logs: logs, startedAt: time.Now()}
}

// Stats 返回仪表盘数据。
func (s *DashboardService) Stats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{GoVersion: runtime.Version()}
	var err error
	stats.Articles, err = s.articles.Count(ctx, "")
	if err != nil {
		return nil, err
	}
	stats.Published, err = s.articles.Count(ctx, string(model.ArticleStatusPublished))
	if err != nil {
		return nil, err
	}
	stats.Users, err = s.users.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Media, err = s.media.Count(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&model.Category{}).Count(&stats.Categories).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&model.Article{}).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.Views).Error; err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	stats.TodayPublished, err = s.articles.CountPublishedSince(ctx, start)
	if err != nil {
		return nil, err
	}
	stats.Trend, err = s.articles.PublishedTrend(ctx, 7)
	if err != nil {
		return nil, err
	}
	stats.Latest, err = s.articles.Recent(ctx, 6)
	if err != nil {
		return nil, err
	}
	stats.Logs, err = s.logs.Latest(ctx, 8)
	if err != nil {
		return nil, err
	}
	stats.Uptime = time.Since(s.startedAt).Truncate(time.Second).String()
	stats.DBVersion = s.dbVersion(ctx)
	return stats, nil
}

func (s *DashboardService) dbVersion(ctx context.Context) string {
	query := "SELECT sqlite_version()"
	if s.db.Dialector.Name() == "mysql" {
		query = "SELECT VERSION()"
	}
	var version string
	if err := s.db.WithContext(ctx).Raw(query).Scan(&version).Error; err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s (%s)", s.db.Dialector.Name(), version)
}
