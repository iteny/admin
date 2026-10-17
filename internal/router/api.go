package router

import "github.com/gin-gonic/gin"

func registerAPI(engine *gin.Engine, deps *Deps) {
	api := engine.Group("/api/v1")
	{
		api.GET("/site", deps.API.Site)
		api.GET("/home", deps.API.Home)
		api.GET("/articles", deps.API.Articles)
		api.GET("/articles/id/:id", deps.API.ArticleByID)
		api.GET("/articles/:slug", deps.API.Article)
		api.GET("/search", deps.API.Articles)
		api.GET("/categories", deps.API.Categories)
		api.GET("/tags", deps.API.Tags)
	}
}
