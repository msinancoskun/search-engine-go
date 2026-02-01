package main

import (
	"html/template"
	"net/http"

	"search-engine-go/internal/api/middleware"
	"search-engine-go/internal/config"

	"github.com/gin-gonic/gin"
)

func setupRouter(cfg *config.Config, deps *Dependencies) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(deps.Logger))
	router.Use(middleware.Recovery(deps.Logger))
	router.Use(middleware.CORS())
	router.Use(deps.RateLimiter.Middleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/login", deps.AuthHandler.LoginPage)
	
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", deps.AuthHandler.Login)
	}

	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTAuth(deps.JWTService, deps.Logger))
	{
		v1.POST("/auth/logout", deps.AuthHandler.Logout)
		
		v1.GET("/search", deps.ContentHandler.Search)
		v1.GET("/content/:id", deps.ContentHandler.GetByID)
	}
	
	docs := router.Group("/docs")
	docs.Use(middleware.JWTAuthHTML(deps.JWTService, deps.Logger))
	{
		docs.StaticFile("/openapi.yaml", "./openapi.yaml")
		docs.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "swagger.html", nil)
		})
	}

	router.Static("/static", "./web/static")
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	})
	router.LoadHTMLGlob("web/templates/*")
	
	dashboard := router.Group("")
	dashboard.Use(middleware.JWTAuthHTML(deps.JWTService, deps.Logger))
	{
		dashboard.GET("/", deps.DashboardHandler.Index)
		dashboard.GET("/dashboard", deps.DashboardHandler.Index)
	}

	return router
}
