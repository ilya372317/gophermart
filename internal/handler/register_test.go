package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	type want struct {
		code int
	}
	type prepareData struct {
		createdUser             *entity.User
		userExists              bool
		hasUserReturnErr        bool
		saveReturnErr           bool
		getUserByLoginReturnErr bool
	}
	type requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	tests := []struct {
		name        string
		prepareData prepareData
		requestBody requestBody
		want        want
	}{
		{
			name: "user with login already registered",
			prepareData: prepareData{
				userExists:       true,
				hasUserReturnErr: false,
				createdUser:      nil,
				saveReturnErr:    false,
			},
			requestBody: requestBody{
				Login:    "test",
				Password: "123",
			},
			want: want{
				code: http.StatusConflict,
			},
		},
		{
			name: "has user function return error",
			prepareData: prepareData{
				userExists:       false,
				hasUserReturnErr: true,
				createdUser:      nil,
				saveReturnErr:    false,
			},
			requestBody: requestBody{
				Login:    "test",
				Password: "123",
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "given empty body",
			prepareData: prepareData{
				userExists:       true,
				hasUserReturnErr: false,
				createdUser:      nil,
				saveReturnErr:    false,
			},
			requestBody: requestBody{
				Login:    "",
				Password: "",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "success register user",
			prepareData: prepareData{
				userExists:       false,
				hasUserReturnErr: false,
				createdUser: &entity.User{
					Login:    "test",
					Password: "123",
					ID:       0,
				},
				saveReturnErr:           false,
				getUserByLoginReturnErr: false,
			},
			requestBody: requestBody{
				Login:    "test",
				Password: "123",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "save user return error",
			prepareData: prepareData{
				userExists:              false,
				hasUserReturnErr:        false,
				createdUser:             nil,
				saveReturnErr:           true,
				getUserByLoginReturnErr: false,
			},
			requestBody: requestBody{
				Login:    "123",
				Password: "123",
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "get user by login return error",
			prepareData: prepareData{
				userExists:       false,
				hasUserReturnErr: false,
				createdUser: &entity.User{
					Login:    "123",
					Password: "123",
					ID:       1,
				},
				saveReturnErr:           false,
				getUserByLoginReturnErr: true,
			},
			requestBody: requestBody{
				Login:    "123",
				Password: "123",
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := handler_mock.NewMockRegisterStorage(ctrl)
			switch {
			case tt.prepareData.hasUserReturnErr:
				repoMock.EXPECT().
					HasUser(ctx, gomock.Eq(tt.requestBody.Login)).
					Return(false, fmt.Errorf("failed check user")).
					AnyTimes()
			case tt.prepareData.userExists:
				repoMock.EXPECT().
					HasUser(ctx, gomock.Eq(tt.requestBody.Login)).
					Return(true, nil).
					AnyTimes()
			default:
				repoMock.EXPECT().
					HasUser(ctx, gomock.Eq(tt.requestBody.Login)).
					Return(false, nil).
					AnyTimes()
			}

			if tt.prepareData.saveReturnErr {
				user := entity.User{
					Login: tt.requestBody.Login,
				}
				user.SetPassword(tt.requestBody.Password)
				repoMock.EXPECT().
					SaveUser(ctx, user).
					Return(fmt.Errorf("failed save user")).
					AnyTimes()
			} else {
				user := entity.User{
					Login: tt.requestBody.Login,
				}
				user.SetPassword(tt.requestBody.Password)
				repoMock.EXPECT().
					SaveUser(ctx, user).
					Return(nil).
					AnyTimes()
			}

			if tt.prepareData.getUserByLoginReturnErr {
				repoMock.EXPECT().
					GetUserByLogin(ctx, gomock.Eq(tt.requestBody.Login)).
					Return(tt.prepareData.createdUser, fmt.Errorf("failed get user by login")).
					AnyTimes()
			} else {
				repoMock.EXPECT().
					GetUserByLogin(ctx, gomock.Eq(tt.requestBody.Login)).
					Return(tt.prepareData.createdUser, nil).
					AnyTimes()
			}

			body, err := json.Marshal(&tt.requestBody)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "localhost:8080/register", bytes.NewReader(body))
			writer := httptest.NewRecorder()
			handler := Register(repoMock)
			handler.ServeHTTP(writer, request)

			res := writer.Result()
			defer func() {
				err = res.Body.Close()
				require.NoError(t, err)
			}()
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				authToken := getAuthTokenFromResponse(res)
				assert.NotEqual(t, "", authToken)
			}
		})
	}
}

func getAuthTokenFromResponse(res *http.Response) string {
	for _, cookie := range res.Cookies() {
		if cookie.Name == "AUTH_TOKEN" {
			return cookie.Name
		}
	}

	return ""
}
