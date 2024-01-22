package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrderList(t *testing.T) {
	now := time.Time{}
	const (
		userID      = uint(1)
		userBalance = 10
	)
	type want struct {
		statusCode     int
		responseStruct []OrderResponse
	}
	tests := []struct {
		name                          string
		getOrderListByUserIDReturnErr bool
		ordersFromStorage             []entity.Order
		want                          want
	}{
		{
			name:                          "success filled case",
			getOrderListByUserIDReturnErr: false,
			ordersFromStorage: []entity.Order{
				{
					CreatedAT: now,
					UpdatedAT: now,
					Status:    entity.StatusInvalid,
					Number:    123,
					Accrual: sql.NullInt64{
						Int64: 123,
						Valid: true,
					},
					ID:     1,
					UserID: userID,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				responseStruct: []OrderResponse{
					{
						Number:     "123",
						Status:     entity.StatusInvalid,
						Accrual:    123,
						UploadedAt: now,
					},
				},
			},
		},
		{
			name:                          "get order return err case",
			getOrderListByUserIDReturnErr: true,
			ordersFromStorage:             nil,
			want: want{
				statusCode:     http.StatusInternalServerError,
				responseStruct: nil,
			},
		},
		{
			name:                          "empty storage case",
			getOrderListByUserIDReturnErr: false,
			ordersFromStorage:             nil,
			want: want{
				statusCode:     http.StatusNoContent,
				responseStruct: nil,
			},
		},
	}
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	authUser := &entity.User{
		CreatedAT: time.Now(),
		UpdatedAT: time.Now(),
		Login:     "123",
		Password:  "123",
		ID:        userID,
		Balance:   userBalance,
	}
	ctxWithUser := context.WithValue(ctx, entity.UserKey, authUser)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := handler_mock.NewMockOrderListStorage(ctrl)
			if tt.getOrderListByUserIDReturnErr {
				repo.EXPECT().GetOrderListByUserID(ctxWithUser, userID).
					Return(nil, fmt.Errorf("failed get order list")).
					MinTimes(1)
			} else {
				repo.EXPECT().GetOrderListByUserID(ctxWithUser, userID).
					Return(tt.ordersFromStorage, nil).
					MinTimes(1)
			}
			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			request = request.WithContext(ctxWithUser)
			writer := httptest.NewRecorder()
			handler := GetOrderList(repo)
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			orderResponse := make([]OrderResponse, 0, len(tt.ordersFromStorage))
			if res.Header.Get("Content-Type") == "application/json" {
				err := json.NewDecoder(res.Body).Decode(&orderResponse)
				require.NoError(t, err)
				assert.Equal(t, tt.want.responseStruct, orderResponse)
			}
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			_ = res.Body.Close()
		})
	}
}
