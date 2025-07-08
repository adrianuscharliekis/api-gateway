package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("mysecret")

func GenerateJWT(sub string) (string, error) {
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GenerateRefreshJWT(sub string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  sub,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // refresh token valid 7 hari
		"type": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func VerifyJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
}

func ExtractClaims(c *gin.Context) (jwt.MapClaims, error) {
	tokenStr := c.GetHeader("Authorization")
	token, err := VerifyJWT(tokenStr)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
