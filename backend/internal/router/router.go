package router

import (
	"github.com/gin-gonic/gin"

	"backend/internal/handler"
	"backend/internal/middleware"
)

func Setup(userHandler *handler.UserHandler) *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok"})
	})

	// API 路由组
	api := r.Group("/api")
	{
		api.POST("/users", userHandler.Create)
		api.GET("/users", userHandler.List)
		api.GET("/users/:id", userHandler.Get)
		api.PUT("/users/:id", userHandler.Update)
		api.DELETE("/users/:id", userHandler.Delete)
	}

	return r
}
