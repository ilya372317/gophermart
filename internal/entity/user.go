package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const nullUserID = 0

type User struct {
	CreatedAT time.Time `db:"created_at"`
	UpdatedAT time.Time `db:"updated_at"`
	Login     string    `db:"login"`
	Password  string    `db:"password"`
	ID        uint      `db:"id"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

func (u *User) SetPassword(pass string) {
	hash := sha256.Sum256([]byte(pass))
	u.Password = hex.EncodeToString(hash[:])
}

func (u *User) GetPasswordHash() string {
	return u.Password
}

func (u *User) CheckPassword(pass string) bool {
	hash := sha256.Sum256([]byte(pass))
	return u.Password == hex.EncodeToString(hash[:])
}

func (u *User) GenerateJWTToken(secretKey string, expTime time.Duration) (string, error) {
	if u.ID == nullUserID {
		return "", fmt.Errorf("for generate jwt token user must have id")
	}
	if secretKey == "" {
		return "", fmt.Errorf("secret key can`t be empty")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expTime)),
		},
		UserID: u.ID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", fmt.Errorf("failed create sign: %w", err)
	}

	return tokenString, nil
}
