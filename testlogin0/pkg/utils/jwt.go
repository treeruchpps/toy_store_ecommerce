package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateToken(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   userID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

func VerifyToken(tokenString string, secret string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า token valid หรือไม่
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

func ParseToken(tokenString string, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims.Subject, nil // คืนค่า userID (หรือ subject) จาก JWT
	}

	return "", err
}
