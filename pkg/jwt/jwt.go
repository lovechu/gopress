package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// Config JWT 配置
type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpire  int
	RefreshExpire int
}

// Claims 自定义声明
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 access token
func GenerateAccessToken(cfg Config, userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.AccessExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gopress",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.AccessSecret))
}

// GenerateRefreshToken 生成 refresh token
func GenerateRefreshToken(cfg Config, userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.RefreshExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gopress",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.RefreshSecret))
}

// ParseAccessToken 解析 access token
func ParseAccessToken(tokenStr, secret string) (*Claims, error) {
	return parseToken(tokenStr, secret, "access")
}

// ParseRefreshToken 解析 refresh token
func ParseRefreshToken(tokenStr, secret string) (*Claims, error) {
	return parseToken(tokenStr, secret, "refresh")
}

func parseToken(tokenStr, secret, tokenType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	if claims.Type != tokenType {
		return nil, fmt.Errorf("%w: wrong token type", ErrTokenInvalid)
	}
	return claims, nil
}
