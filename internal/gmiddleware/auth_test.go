package gmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/auth"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/entity"
	gmiddleware_mock "github.com/ilya372317/gophermart/internal/gmiddleware/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	tests := []struct {
		name          string
		clientKey     string
		userID        uint
		want          int
		hasAuthCookie bool
		repoHasUser   bool
	}{
		{
			name:          "success check auth case",
			hasAuthCookie: true,
			userID:        1,
			repoHasUser:   true,
			want:          http.StatusOK,
			clientKey:     "secret-key",
		},
		{
			name:          "not have cookie",
			hasAuthCookie: false,
			userID:        1,
			repoHasUser:   true,
			want:          http.StatusUnauthorized,
			clientKey:     "secret-key",
		},
		{
			name:          "storage not contain user",
			hasAuthCookie: true,
			userID:        1,
			repoHasUser:   false,
			want:          http.StatusInternalServerError,
			clientKey:     "secret-key",
		},
		{
			name:          "invalid token case",
			hasAuthCookie: true,
			userID:        1,
			repoHasUser:   true,
			want:          http.StatusUnauthorized,
			clientKey:     "invalid-key",
		},
	}
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	gopherConfig := &config.GophermartConfig{
		SecretKey: "secret-key",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &entity.User{
				Login: "login",
				ID:    tt.userID,
			}
			repo := gmiddleware_mock.NewMockAuthStorage(ctrl)
			if tt.repoHasUser {
				repo.EXPECT().GetUserByID(ctx, gomock.Eq(tt.userID)).
					Return(user, nil).
					AnyTimes()
			} else {
				repo.EXPECT().GetUserByID(ctx, gomock.Eq(tt.userID)).
					Return(nil, fmt.Errorf("failed get user by id")).
					AnyTimes()
			}
			request := httptest.NewRequest(http.MethodPost, "/test-route", nil)
			if tt.hasAuthCookie {
				user.SetPassword("password")
				authService := auth.New(tt.clientKey, time.Second)
				token, err := authService.GenerateJWTToken(user)
				require.NoError(t, err)
				request.AddCookie(&http.Cookie{
					Name:    "AUTH_TOKEN",
					Value:   token,
					Expires: time.Time{},
				})
			}
			writer := httptest.NewRecorder()
			handlerFunc := Auth(gopherConfig, repo)
			handler := handlerFunc(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			err := res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
