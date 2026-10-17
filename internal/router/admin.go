package router

import (
	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
)

func registerAdmin(engine *gin.Engine, deps *Deps) {
	public := engine.Group("/admin")
	{
		public.GET("/login", deps.Auth.LoginPage)
		public.POST("/login", deps.Auth.Login)
		public.POST("/logout", deps.Auth.Logout)
	}

	admin := engine.Group("/admin")
	admin.Use(middleware.AuthRequired(deps.AuthService), middleware.AdminMenus(deps.MenuService))
	{
		admin.GET("", middleware.Permission("dashboard:view"), deps.Dashboard.Index)
		admin.GET("/api/dashboard/stats", middleware.Permission("dashboard:view"), deps.Dashboard.StatsJSON)

		admin.GET("/articles", middleware.Permission("article:view"), deps.Articles.Index)
		admin.GET("/articles/create", middleware.Permission("article:create"), deps.Articles.CreatePage)
		admin.POST("/articles", middleware.Permission("article:create"), deps.Articles.Create)
		admin.GET("/articles/:id/edit", middleware.Permission("article:update"), deps.Articles.EditPage)
		admin.POST("/articles/:id", middleware.Permission("article:update"), deps.Articles.Update)
		admin.PUT("/articles/:id", middleware.Permission("article:update"), deps.Articles.Update)
		admin.POST("/articles/:id/delete", middleware.Permission("article:delete"), deps.Articles.Delete)
		admin.DELETE("/articles/:id/delete", middleware.Permission("article:delete"), deps.Articles.Delete)
		admin.POST("/articles/batch", middleware.RequireAnyPermission("article:update", "article:delete"), deps.Articles.Batch)
		admin.GET("/articles/:id/preview", middleware.Permission("article:view"), deps.Articles.Preview)

		admin.GET("/categories", middleware.Permission("category:view"), deps.Categories.Index)
		admin.POST("/categories", middleware.Permission("category:create"), deps.Categories.Create)
		admin.POST("/categories/reorder", middleware.Permission("category:update"), deps.Categories.Reorder)
		admin.POST("/categories/:id", middleware.Permission("category:update"), deps.Categories.Update)
		admin.PUT("/categories/:id", middleware.Permission("category:update"), deps.Categories.Update)
		admin.POST("/categories/:id/delete", middleware.Permission("category:delete"), deps.Categories.Delete)
		admin.DELETE("/categories/:id/delete", middleware.Permission("category:delete"), deps.Categories.Delete)

		admin.GET("/tags", middleware.Permission("tag:view"), deps.Tags.Index)
		admin.POST("/tags", middleware.Permission("tag:create"), deps.Tags.Create)
		admin.POST("/tags/:id", middleware.Permission("tag:update"), deps.Tags.Update)
		admin.PUT("/tags/:id", middleware.Permission("tag:update"), deps.Tags.Update)
		admin.POST("/tags/:id/delete", middleware.Permission("tag:delete"), deps.Tags.Delete)
		admin.DELETE("/tags/:id/delete", middleware.Permission("tag:delete"), deps.Tags.Delete)

		admin.GET("/pages", middleware.Permission("page:view"), deps.Pages.Index)
		admin.GET("/pages/create", middleware.Permission("page:create"), deps.Pages.CreatePage)
		admin.POST("/pages", middleware.Permission("page:create"), deps.Pages.Create)
		admin.GET("/pages/:id/edit", middleware.Permission("page:update"), deps.Pages.EditPage)
		admin.POST("/pages/:id", middleware.Permission("page:update"), deps.Pages.Update)
		admin.PUT("/pages/:id", middleware.Permission("page:update"), deps.Pages.Update)
		admin.POST("/pages/:id/delete", middleware.Permission("page:delete"), deps.Pages.Delete)
		admin.DELETE("/pages/:id/delete", middleware.Permission("page:delete"), deps.Pages.Delete)

		admin.GET("/media", middleware.Permission("media:view"), deps.Media.Index)
		admin.POST("/media/upload", middleware.Permission("media:upload"), deps.Media.Upload)
		admin.GET("/media/picker", middleware.Permission("media:view"), deps.Media.Picker)
		admin.POST("/media/:id/delete", middleware.Permission("media:delete"), deps.Media.Delete)
		admin.DELETE("/media/:id/delete", middleware.Permission("media:delete"), deps.Media.Delete)

		admin.GET("/users", middleware.Permission("user:view"), deps.Users.Index)
		admin.GET("/users/create", middleware.Permission("user:create"), deps.Users.CreatePage)
		admin.POST("/users", middleware.Permission("user:create"), deps.Users.Create)
		admin.GET("/users/:id/edit", middleware.Permission("user:update"), deps.Users.EditPage)
		admin.POST("/users/:id", middleware.Permission("user:update"), deps.Users.Update)
		admin.PUT("/users/:id", middleware.Permission("user:update"), deps.Users.Update)
		admin.POST("/users/:id/delete", middleware.Permission("user:delete"), deps.Users.Delete)
		admin.DELETE("/users/:id/delete", middleware.Permission("user:delete"), deps.Users.Delete)
		admin.POST("/users/:id/reset-password", middleware.Permission("user:reset_password"), deps.Users.ResetPassword)

		admin.GET("/roles", middleware.Permission("role:view"), deps.Roles.Index)
		admin.GET("/roles/create", middleware.Permission("role:create"), deps.Roles.CreatePage)
		admin.POST("/roles", middleware.Permission("role:create"), deps.Roles.Create)
		admin.GET("/roles/:id/edit", middleware.Permission("role:update"), deps.Roles.EditPage)
		admin.POST("/roles/:id", middleware.Permission("role:update"), deps.Roles.Update)
		admin.PUT("/roles/:id", middleware.Permission("role:update"), deps.Roles.Update)
		admin.POST("/roles/:id/delete", middleware.Permission("role:delete"), deps.Roles.Delete)
		admin.DELETE("/roles/:id/delete", middleware.Permission("role:delete"), deps.Roles.Delete)

		admin.GET("/menus", middleware.Permission("menu:view"), deps.Menus.Index)
		admin.POST("/menus", middleware.Permission("menu:create"), deps.Menus.Create)
		admin.POST("/menus/reorder", middleware.Permission("menu:update"), deps.Menus.Reorder)
		admin.POST("/menus/:id", middleware.Permission("menu:update"), deps.Menus.Update)
		admin.PUT("/menus/:id", middleware.Permission("menu:update"), deps.Menus.Update)
		admin.POST("/menus/:id/delete", middleware.Permission("menu:delete"), deps.Menus.Delete)
		admin.DELETE("/menus/:id/delete", middleware.Permission("menu:delete"), deps.Menus.Delete)

		admin.GET("/settings", middleware.Permission("setting:view"), deps.Settings.Index)
		admin.POST("/settings", middleware.Permission("setting:update"), deps.Settings.Save)

		admin.GET("/logs", middleware.Permission("log:view"), deps.Logs.Index)
		admin.POST("/logs/clear", middleware.Permission("log:delete"), deps.Logs.Clear)

		admin.GET("/profile", deps.Profile.Index)
		admin.POST("/profile", deps.Profile.Update)
		admin.POST("/profile/password", deps.Profile.Password)
	}
}
