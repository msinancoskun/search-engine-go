package middleware

import (
	"net/http"
	"strings"

	"search-engine-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func JWTAuth(jwtService *service.JWTService, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing authorization header", 
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Authorization header is required. Please provide a valid JWT token.",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid authorization header format",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header",
				"message": "Authorization header must be in the format: 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			log.Warn("Empty token in authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token is missing or empty",
			})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Warn("Token validation failed", 
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"message": "The provided token is invalid, expired, or malformed. Please login again.",
			})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

func JWTAuthHTML(jwtService *service.JWTService, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookie, err := c.Cookie("jwt_token")
			if err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			log.Warn("Missing JWT token for HTML route", zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Warn("Token validation failed for HTML route", zap.Error(err), zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
