package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cms/internal/middleware"
	"cms/internal/pkg/response"
)

func registerWeb(engine *gin.Engine, deps *Deps) {
	front := engine.Group("")
	front.Use(middleware.FrontNav(deps.Front))
	{
		front.GET("/", deps.Home.Index)
		front.GET("/page/:page", deps.Home.Index)
		front.GET("/category/:slug", deps.WebCategory.Show)
		front.GET("/category/:slug/page/:page", deps.WebCategory.Show)
		front.GET("/article/:slug", deps.WebArticle.Show)
		front.GET("/tag/:slug", deps.WebTag.Show)
		front.GET("/tag/:slug/page/:page", deps.WebTag.Show)
		front.GET("/p/:slug", deps.WebPage.Show)
		front.GET("/search", deps.Search.Index)
		front.GET("/sitemap.xml", deps.SEO.Sitemap)
		front.GET("/robots.txt", deps.SEO.Robots)
	}

	engine.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			response.Fail(c, http.StatusNotFound, 40400, "接口不存在")
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.HTML(http.StatusNotFound, "admin/errors/404.html", gin.H{
				"Title":     "页面不存在",
				"SiteName":  middleware.SiteName(c),
				"CSRFToken": middleware.Token(c),
				"User":      middleware.CurrentUser(c),
				"Perms":     middleware.Permissions(c),
				"Menus":     middleware.AdminMenuTree(c),
				"Flash":     middleware.Messages(c),
			})
			return
		}
		c.HTML(http.StatusNotFound, "errors/404.html", gin.H{
			"Title":    "页面不存在",
			"SiteName": middleware.SiteName(c),
			"Message":  "你访问的页面不存在或已被移动。",
		})
	})
}
