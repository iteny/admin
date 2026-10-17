package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"cms/internal/model"
)

// startLogCleanup 每天清理 90 天前的操作日志。
func startLogCleanup(db *gorm.DB, logger *slog.Logger) {
	go func() {
		cleanup := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cutoff := time.Now().UTC().AddDate(0, 0, -90)
			if err := db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&model.OperationLog{}).Error; err != nil {
				logger.Warn("cleanup operation logs failed", "error", err)
			}
		}
		cleanup()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanup()
		}
	}()
}
