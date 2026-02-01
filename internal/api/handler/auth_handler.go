package handler

import (
	"net/http"

	"search-engine-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	jwtService *service.JWTService
	log        *zap.Logger
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(jwtService *service.JWTService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
		log:        log,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// TODO: Implement proper user authentication against database
	if req.Username == "admin" && req.Password == "admin" {
		token, err := h.jwtService.GenerateToken(req.Username)
		if err != nil {
			h.log.Error("Failed to generate token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{Token: token})
		return
	}

	h.log.Warn("Invalid credentials", zap.String("username", req.Username))
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}
