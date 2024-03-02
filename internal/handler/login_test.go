package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dto"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name                    string
		body                    dto.UserCredentials
		wantCode                int
		createdUser             *entity.User
		getUserByLoginReturnErr bool
		sqlNoRows               bool
		secretKey               string
	}{
		{
			name: "success case",
			body: dto.UserCredentials{
				Login:    "test-login",
				Password: "123",
			},
			wantCode: http.StatusOK,
			createdUser: &entity.User{
				Login:    "test-login",
				Password: "123",
				ID:       1,
			},
			getUserByLoginReturnErr: false,
			secretKey:               "secret-key",
			sqlNoRows:               false,
		},
		{
			name: "invalid credentials",
			body: dto.UserCredentials{
				Login:    "",
				Password: "",
			},
			wantCode:                http.StatusBadRequest,
			createdUser:             nil,
			getUserByLoginReturnErr: false,
			secretKey:               "secret-key",
			sqlNoRows:               false,
		},
		{
			name: "get user by login return err",
			body: dto.UserCredentials{
				Login:    "123",
				Password: "123",
			},
			wantCode:                http.StatusInternalServerError,
			createdUser:             nil,
			getUserByLoginReturnErr: true,
			secretKey:               "secret-key",
			sqlNoRows:               false,
		},
		{
			name: "given password invalid",
			body: dto.UserCredentials{
				Login:    "test-login",
				Password: "123",
			},
			wantCode: http.StatusUnauthorized,
			createdUser: &entity.User{
				Login:    "test-login",
				Password: "321",
				ID:       1,
			},
			secretKey:               "secret-key",
			getUserByLoginReturnErr: false,
			sqlNoRows:               false,
		},
		{
			name: "failed generate jwt token",
			body: dto.UserCredentials{
				Login:    "123",
				Password: "123",
			},
			wantCode: http.StatusInternalServerError,
			createdUser: &entity.User{
				Login:    "123",
				Password: "123",
				ID:       1,
			},
			getUserByLoginReturnErr: false,
			secretKey:               "",
		},
		{
			name: "not existing login",
			body: dto.UserCredentials{
				Login:    "teset-123",
				Password: "test321",
			},
			wantCode:                http.StatusUnauthorized,
			createdUser:             nil,
			getUserByLoginReturnErr: true,
			secretKey:               "secret-key",
			sqlNoRows:               true,
		},
	}
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strg := handler_mock.NewMockLoginStorage(ctrl)
			if tt.createdUser != nil {
				tt.createdUser.SetPassword(tt.createdUser.Password)
			}
			if tt.getUserByLoginReturnErr {
				if tt.sqlNoRows {
					strg.EXPECT().
						GetUserByLogin(ctx, gomock.Eq(tt.body.Login)).
						Return(nil, sql.ErrNoRows).
						AnyTimes()
				} else {
					strg.EXPECT().
						GetUserByLogin(ctx, gomock.Eq(tt.body.Login)).
						Return(nil, fmt.Errorf("failed get user by login")).
						AnyTimes()
				}
			} else {
				strg.EXPECT().
					GetUserByLogin(ctx, gomock.Eq(tt.body.Login)).
					Return(tt.createdUser, nil).
					AnyTimes()
			}
			cnfg := &config.GophermartConfig{
				SecretKey: tt.secretKey,
				ExpTime:   1,
			}
			bodyData, err := json.Marshal(&tt.body)
			require.NoError(t, err)
			request := httptest.NewRequest(
				http.MethodPost,
				"localhost:8080/login", bytes.NewReader(bodyData))
			writer := httptest.NewRecorder()
			handler := Login(strg, cnfg)
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}
