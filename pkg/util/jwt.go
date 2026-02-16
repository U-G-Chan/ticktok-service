package util

import (
	"time"
	"ticktok-service/pkg/config"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates access and refresh tokens
func GenerateToken(userID int64) (string, string, error) {
	// Access Token
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.JWT.AccessExpire) * time.Second)),
			Issuer:    "ticktok",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(config.Config.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.JWT.RefreshExpire) * time.Second)),
			Issuer:    "ticktok",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(config.Config.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

// ParseToken parses the token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
