package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cms/internal/config"
	adminhandler "cms/internal/handler/admin"
	apihandler "cms/internal/handler/api"
	webhandler "cms/internal/handler/web"
	"cms/internal/middleware"
	"cms/internal/render"
	"cms/internal/repository"
	"cms/internal/router"
	"cms/internal/service"
	"cms/web"
)

// App 持有应用运行期依赖。
type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *gorm.DB
	Engine *gin.Engine
	Server *http.Server
	Loc    *time.Location
	closer func()
}

// New 按顺序装配应用。调用方必须在使用完毕后调用 Close。
func New(configPath, version string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", cfg.App.Timezone, err)
	}
	logger, closer, err := NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(logger)

	db, err := OpenDatabase(cfg, logger)
	if err != nil {
		if closer != nil {
			_ = closer.Close()
		}
		return nil, err
	}
	if cfg.App.AutoMigrate {
		if err := Migrate(db); err != nil {
			return nil, err
		}
	}
	if cfg.App.Seed {
		if err := Seed(db, logger); err != nil {
			return nil, err
		}
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if !cfg.App.Debug {
		gin.SetMode(gin.TestMode)
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	permissionRepo := repository.NewPermissionRepo(db)
	menuRepo := repository.NewMenuRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	tagRepo := repository.NewTagRepo(db)
	articleRepo := repository.NewArticleRepo(db)
	pageRepo := repository.NewPageRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	settingRepo := repository.NewSettingRepo(db)
	logRepo := repository.NewLogRepo(db)

	// Services
	settings := service.NewSettingService(settingRepo)
	if err := settings.Load(context.Background()); err != nil {
		return nil, err
	}
	auth := service.NewAuthService(userRepo)
	articles := service.NewArticleService(articleRepo, categoryRepo, tagRepo)
	categories := service.NewCategoryService(categoryRepo)
	tags := service.NewTagService(tagRepo)
	pages := service.NewPageService(pageRepo)
	media := service.NewMediaService(mediaRepo, cfg, settings)
	users := service.NewUserService(userRepo, roleRepo)
	roles := service.NewRoleService(roleRepo, permissionRepo)
	menus := service.NewMenuService(menuRepo)
	dashboard := service.NewDashboardService(db, articleRepo, categoryRepo, userRepo, mediaRepo, logRepo)
	front := service.NewFrontService(articles, categories, tags, pages)

	// Renderer
	templateFS, err := templateFileSystem(cfg)
	if err != nil {
		return nil, err
	}
	renderer, err := render.New(templateFS, cfg.App.Debug, render.FuncMap(loc))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	// Session store
	store := cookie.NewStore(cfg.SessionSecret())
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   cfg.Session.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	// Handlers
	adminBase := adminhandler.NewBase(cfg, settings, menus, logger, loc, version)
	authHandler := adminhandler.NewAuthHandler(adminBase, auth)
	dashboardHandler := adminhandler.NewDashboardHandler(adminBase, dashboard)
	articleHandler := adminhandler.NewArticleHandler(adminBase, articles, categories, tags)
	categoryHandler := adminhandler.NewCategoryHandler(adminBase, categories)
	tagHandler := adminhandler.NewTagHandler(adminBase, tags)
	pageHandler := adminhandler.NewPageHandler(adminBase, pages)
	mediaHandler := adminhandler.NewMediaHandler(adminBase, media)
	userHandler := adminhandler.NewUserHandler(adminBase, users, roles)
	roleHandler := adminhandler.NewRoleHandler(adminBase, roles)
	menuHandler := adminhandler.NewMenuHandler(adminBase, menus, roles)
	settingHandler := adminhandler.NewSettingHandler(adminBase, settings)
	logHandler := adminhandler.NewLogHandler(adminBase, logRepo, users)
	profileHandler := adminhandler.NewProfileHandler(adminBase, users, auth)

	webBase := webhandler.NewBase(cfg, settings, front, logger, loc, version)
	homeHandler := webhandler.NewHomeHandler(webBase)
	webArticleHandler := webhandler.NewArticleHandler(webBase)
	webCategoryHandler := webhandler.NewCategoryHandler(webBase)
	webTagHandler := webhandler.NewTagHandler(webBase)
	webPageHandler := webhandler.NewPageHandler(webBase)
	searchHandler := webhandler.NewSearchHandler(webBase)
	seoHandler := webhandler.NewSEOHandler(webBase)
	apiHandler := apihandler.New(cfg, settings, front, articles, categories, tags, logger, loc, version)

	engine := gin.New()
	engine.HTMLRender = renderer
	engine.Use(
		middleware.Recovery(logger),
		middleware.AccessLog(logger, logRepo),
		sessions.Sessions(cfg.Session.Name, store),
		middleware.InjectSiteSettings(settings),
		middleware.FlashMiddleware(),
		middleware.MethodOverride(),
		middleware.CSRF(),
	)

	// Static assets: embedded in production, disk in debug for live reload.
	registerStatic(engine, cfg)

	if err := os.MkdirAll(filepath.Join("storage", "uploads"), 0o750); err != nil {
		return nil, fmt.Errorf("create uploads dir: %w", err)
	}
	engine.StaticFS("/uploads", gin.Dir(filepath.Join("storage", "uploads"), false))

	router.Register(engine, &router.Deps{
		Auth: authHandler, Dashboard: dashboardHandler, Articles: articleHandler,
		Categories: categoryHandler, Tags: tagHandler, Pages: pageHandler, Media: mediaHandler,
		Users: userHandler, Roles: roleHandler, Menus: menuHandler, Settings: settingHandler,
		Logs: logHandler, Profile: profileHandler,
		Home: homeHandler, WebArticle: webArticleHandler, WebCategory: webCategoryHandler,
		WebTag: webTagHandler, WebPage: webPageHandler, Search: searchHandler, SEO: seoHandler,
		API:         apiHandler,
		AuthService: auth, MenuService: menus, Front: front, SettingsSvc: settings, LogRepo: logRepo, Logger: logger,
	})

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	startLogCleanup(db, logger)

	app := &App{Config: cfg, Logger: logger, DB: db, Engine: engine, Loc: loc}
	if closer != nil {
		app.closer = func() { _ = closer.Close() }
	}
	app.Server = &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	return app, nil
}

func templateFileSystem(cfg *config.Config) (fs.FS, error) {
	if cfg.App.Debug {
		return os.DirFS(filepath.Join("web", "template")), nil
	}
	return fs.Sub(web.FS, "template")
}

func registerStatic(engine *gin.Engine, cfg *config.Config) {
	if cfg.App.Debug {
		engine.StaticFS("/static", http.Dir(filepath.Join("web", "static")))
		return
	}
	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		return
	}
	engine.StaticFS("/static", http.FS(staticFS))
}

// Run 启动 HTTP 服务并等待 ctx 取消后优雅关闭。
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("HTTP server listening", "addr", a.Server.Addr, "env", a.Config.App.Env)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout)
		defer cancel()
		a.Logger.Info("shutting down HTTP server")
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}

// Close 释放数据库与日志资源。
func (a *App) Close() {
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	if a.closer != nil {
		a.closer()
	}
}
