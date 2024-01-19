package gmiddleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/entity"
)

type AuthStorage interface {
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
}

func Auth(gopherConfig *config.GophermartConfig, authStorage AuthStorage) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			tokenString, err := getAuthTokenFromCookie(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return
			}
			claims := &entity.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signed method: %w", err)
				}
				return []byte(gopherConfig.SecretKey), nil
			})
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return
			}
			if !token.Valid {
				http.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := authStorage.GetUserByID(request.Context(), claims.UserID)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(request.Context(), entity.UserKey, user)

			handler.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

func getAuthTokenFromCookie(r *http.Request) (string, error) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == "AUTH_TOKEN" {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("unauthorized")
}
