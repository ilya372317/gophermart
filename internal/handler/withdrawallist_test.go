package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalList(t *testing.T) {
	const userID = uint(1)
	now := time.Now()
	type repoState struct {
		returnErr   bool
		returnValue []entity.Withdrawal
	}
	tests := []struct {
		name      string
		repoState repoState
		want      int
	}{
		{
			name: "success case",
			repoState: repoState{
				returnErr: false,
				returnValue: []entity.Withdrawal{
					{
						CreatedAT:   now,
						UpdatedAT:   now,
						OrderNumber: 123,
						ID:          1,
						UserID:      userID,
						Sum:         123,
					},
					{
						CreatedAT:   now,
						UpdatedAT:   now,
						OrderNumber: 321,
						ID:          2,
						UserID:      userID,
						Sum:         100,
					},
					{
						CreatedAT:   now,
						UpdatedAT:   now,
						OrderNumber: 456,
						ID:          3,
						UserID:      userID,
						Sum:         300,
					},
				},
			},
			want: http.StatusOK,
		},
		{
			name:      "empty case",
			repoState: repoState{},
			want:      http.StatusNoContent,
		},
		{
			name: "get order list return err",
			repoState: repoState{
				returnErr:   true,
				returnValue: nil,
			},
			want: http.StatusInternalServerError,
		},
	}
	ctx := context.WithValue(context.Background(), entity.UserKey, &entity.User{
		ID: userID,
	})
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := handler_mock.NewMockWithdrawalListStorage(ctrl)
			if tt.repoState.returnErr {
				repo.EXPECT().GetWithdrawalListByUserID(ctx, userID).
					Return(nil, fmt.Errorf("failed get withdrawal list")).
					AnyTimes()
			} else {
				repo.EXPECT().GetWithdrawalListByUserID(ctx, userID).
					Return(tt.repoState.returnValue, nil).
					AnyTimes()
			}
			request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			request = request.WithContext(ctx)
			writer := httptest.NewRecorder()
			handler := WithdrawalList(repo)
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			if res.StatusCode == http.StatusOK {
				responseList := make([]WithdrawalResponse, 0, len(tt.repoState.returnValue))
				err := json.NewDecoder(res.Body).Decode(&responseList)
				require.NoError(t, err)
				for _, want := range tt.repoState.returnValue {
					if len(responseList) == 0 {
						t.Error("withdrawals in storage not equals to response item count")
						return
					}
					var got WithdrawalResponse
					got, responseList = responseList[0], responseList[1:]
					assert.Equal(t, want.Sum, got.Sum)
					assert.Equal(t, strconv.Itoa(want.OrderNumber), got.Order)
					assert.Equal(t, want.CreatedAT.Format(time.RFC3339), got.ProcessedAT.Format(time.RFC3339))
				}
			}
			err := res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
