package service

import (
	"errors"
	"time"

	"search-engine-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTService struct {
	secret     []byte
	expiration time.Duration
	log        *zap.Logger
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg config.AuthConfig, log *zap.Logger) *JWTService {
	return &JWTService{
		secret:     []byte(cfg.JWTSecret),
		expiration: cfg.JWTExpiration,
		log:        log,
	}
}

func (s *JWTService) GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(s.expiration)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "search-engine-go",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.log.Error("Failed to generate token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secret, nil
	})

	if err != nil {
		s.log.Warn("Token validation failed", zap.Error(err))
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
