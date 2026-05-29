package router

import (
	"github.com/gin-gonic/gin"

	"backend/internal/handler"
	"backend/internal/middleware"
)

func Setup(userHandler *handler.UserHandler, jwtSecret string) *gin.Engine {
	r := gin.New()

	r.SetTrustedProxies(nil) //默认不信任任何代理

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
		api.GET("/users/:id", middleware.JWTAuth(jwtSecret), userHandler.Get)
		api.PUT("/users/:id", middleware.JWTAuth(jwtSecret), userHandler.Update)
		api.DELETE("/users/:id", userHandler.Delete)
		api.POST("/login", userHandler.Login)
		//api.POST("/activities/:id/register", middleware.JWTAuth(), regHandler.Register)
	}

	return r
}
