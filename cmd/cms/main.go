// Command cms 是 CMS 的唯一入口。
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cms/internal/bootstrap"
)

var version = "1.0.0"

func main() {
	configPath := flag.String("config", "./config/config.yaml", "配置文件路径")
	migrateOnly := flag.Bool("migrate", false, "仅执行数据库迁移后退出")
	seedOnly := flag.Bool("seed", false, "仅初始化种子数据后退出")
	showVersion := flag.Bool("version", false, "打印版本号")
	flag.Parse()

	if *showVersion {
		fmt.Println("cms", version)
		return
	}

	app, err := bootstrap.New(*configPath, version)
	if err != nil {
		slog.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}
	defer app.Close()

	if *migrateOnly {
		if err := bootstrap.Migrate(app.DB); err != nil {
			slog.Error("migrate failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("migration completed")
		return
	}
	if *seedOnly {
		if err := bootstrap.Seed(app.DB, app.Logger); err != nil {
			slog.Error("seed failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("seed completed")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
