package bootstrap

import (
	"log/slog"

	"gorm.io/gorm"

	"cms/internal/seed"
)

// Seed 调用种子初始化。
func Seed(db *gorm.DB, logger *slog.Logger) error {
	return seed.Run(db, logger)
}
