// Package router 负责注册后台与前台路由。
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	adminhandler "cms/internal/handler/admin"
	apihandler "cms/internal/handler/api"
	webhandler "cms/internal/handler/web"
	"cms/internal/repository"
	"cms/internal/service"
)

// Deps 是路由注册所需的 handler 与中间件依赖。
type Deps struct {
	Auth       *adminhandler.AuthHandler
	Dashboard  *adminhandler.DashboardHandler
	Articles   *adminhandler.ArticleHandler
	Categories *adminhandler.CategoryHandler
	Tags       *adminhandler.TagHandler
	Pages      *adminhandler.PageHandler
	Media      *adminhandler.MediaHandler
	Users      *adminhandler.UserHandler
	Roles      *adminhandler.RoleHandler
	Menus      *adminhandler.MenuHandler
	Settings   *adminhandler.SettingHandler
	Logs       *adminhandler.LogHandler
	Profile    *adminhandler.ProfileHandler

	Home        *webhandler.HomeHandler
	WebArticle  *webhandler.ArticleHandler
	WebCategory *webhandler.CategoryHandler
	WebTag      *webhandler.TagHandler
	WebPage     *webhandler.PageHandler
	Search      *webhandler.SearchHandler
	SEO         *webhandler.SEOHandler
	API         *apihandler.Handler

	AuthService *service.AuthService
	MenuService *service.MenuService
	Front       *service.FrontService
	SettingsSvc *service.SettingService
	LogRepo     *repository.LogRepo
	Logger      *slog.Logger
}

// Register 注册全部路由。
func Register(engine *gin.Engine, deps *Deps) {
	registerAPI(engine, deps)
	registerAdmin(engine, deps)
	registerWeb(engine, deps)
}
