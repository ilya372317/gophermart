package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ilya372317/gophermart/internal/entity"
)

const nullUserID = 0

type Service struct {
	secretKey string
	expTime   time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

func New(secretKey string, expTime time.Duration) *Service {
	return &Service{secretKey: secretKey, expTime: expTime}
}

func (s *Service) GenerateJWTToken(user *entity.User) (string, error) {
	if user.ID == nullUserID {
		return "", fmt.Errorf("for generate jwt token user must have id")
	}
	if s.secretKey == "" {
		return "", fmt.Errorf("secret key can`t be empty")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expTime)),
		},
		UserID: user.ID,
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))

	if err != nil {
		return "", fmt.Errorf("failed create sign: %w", err)
	}

	return tokenString, nil
}
