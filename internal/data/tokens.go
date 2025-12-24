package data

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// TokenModel handles JWT token operations
type TokenModel struct {
	Secret string
}

// Claims represents JWT claims
type Claims struct {
	AuthUserID int64  `json:"auth_user_id"`
	Login      string `json:"login"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func (m TokenModel) GenerateToken(authUserID int64, login string, duration time.Duration) (string, error) {
	claims := Claims{
		AuthUserID: authUserID,
		Login:      login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.Secret))
}

// ValidateToken validates a JWT token and returns the claims
func (m TokenModel) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
