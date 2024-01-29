package handler

import (
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

func TestGetUserBalance(t *testing.T) {
	type repoReturnValue struct {
		value float64
		err   error
	}
	type want struct {
		statusCode int
		response   UserBalanceResponse
	}
	tests := []struct {
		name            string
		authUser        *entity.User
		repoReturnValue repoReturnValue
		want            want
	}{
		{
			name: "success case",
			authUser: &entity.User{
				ID:      1,
				Balance: 10,
			},
			repoReturnValue: repoReturnValue{
				value: 50,
				err:   nil,
			},
			want: want{
				statusCode: http.StatusOK,
				response: UserBalanceResponse{
					Current:   10,
					Withdrawn: 50,
				},
			},
		},
		{
			name: "get withdrawals sum return err",
			authUser: &entity.User{
				ID:      1,
				Balance: 10,
			},
			repoReturnValue: repoReturnValue{
				value: 0,
				err:   fmt.Errorf("failed get data"),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   UserBalanceResponse{},
			},
		},
	}
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := context.WithValue(ctx, entity.UserKey, tt.authUser)
			repo := handler_mock.NewMockUserBalanceStorage(ctrl)
			repo.
				EXPECT().
				GetWithdrawalSumByUserID(userCtx, tt.authUser.ID).
				Return(tt.repoReturnValue.value, tt.repoReturnValue.err).
				AnyTimes()

			handler := GetUserBalance(repo)
			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			request = request.WithContext(userCtx)
			writer := httptest.NewRecorder()
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				gotResponse := UserBalanceResponse{}
				err := json.NewDecoder(res.Body).Decode(&gotResponse)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, gotResponse)
			}
		})
	}
}
