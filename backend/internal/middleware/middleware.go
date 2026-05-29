package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"backend/pkg/logger"
	"backend/pkg/response"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Log.Info("request",
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func JWTAuth(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		//fmt.Println("Authorization", auth)
		if auth == "" {
			response.Error(ctx, http.StatusUnauthorized, "missing authorization header")
			ctx.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(ctx, http.StatusUnauthorized, "invalid token")
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(ctx, http.StatusUnauthorized, "invalid claims")
			ctx.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			response.Error(ctx, http.StatusUnauthorized, "invalid user_id in token")
			ctx.Abort()
			return
		}

		ctx.Set("userID", uint(userID))
		ctx.Next()
	}
}
