package bootstrap

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"cms/internal/config"
	"cms/internal/model"
)

// OpenDatabase 根据配置打开数据库并设置连接池。
func OpenDatabase(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Database.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.Database.DSN)
	case "sqlite":
		if cfg.Database.DSN != ":memory:" {
			if err := os.MkdirAll(filepath.Dir(cfg.Database.DSN), 0o750); err != nil {
				return nil, fmt.Errorf("create sqlite dir: %w", err)
			}
		}
		dialector = sqlite.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Database.Driver)
	}

	level := gormlogger.Warn
	if cfg.App.Debug {
		level = gormlogger.Info
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "cms_"},
		Logger:         gormlogger.Default.LogMode(level),
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.Database.MaxLifetime)
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if cfg.Database.Driver == "sqlite" {
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			log.Warn("enable sqlite foreign keys failed", "error", err)
		}
	}
	return db, nil
}

// Migrate 执行 AutoMigrate。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	// 文章状态已收敛为发布/归档，历史草稿统一转为归档，避免前台误发布。
	if err := db.Model(&model.Article{}).
		Where("status = ?", model.ArticleStatusDraft).
		Update("status", model.ArticleStatusArchived).Error; err != nil {
		return fmt.Errorf("normalize article status: %w", err)
	}
	// 单页状态同步收敛为发布/归档。
	if err := db.Model(&model.Page{}).
		Where("status = ?", model.PageStatusDraft).
		Update("status", model.PageStatusArchived).Error; err != nil {
		return fmt.Errorf("normalize page status: %w", err)
	}
	return nil
}
