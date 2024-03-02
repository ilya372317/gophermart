package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawBonus(t *testing.T) {
	const validLuhnString = "12345678903"
	defaultAuthUser := &entity.User{
		Login:    "123",
		Password: "123",
		ID:       1,
		Balance:  100,
	}
	type repoState struct {
		updateUserBalanceReturnErr bool
		saveWithdrawalReturnErr    bool
	}
	type body struct {
		Order string  `json:"order,omitempty"`
		Sum   float64 `json:"sum,omitempty"`
	}
	tests := []struct {
		repoState repoState
		authUser  *entity.User
		name      string
		body      body
		want      int
	}{
		{
			repoState: repoState{
				updateUserBalanceReturnErr: false,
				saveWithdrawalReturnErr:    false,
			},
			authUser: defaultAuthUser,
			name:     "success case",
			body: body{
				Order: validLuhnString,
				Sum:   50,
			},
			want: http.StatusOK,
		},
		{
			repoState: repoState{},
			authUser:  defaultAuthUser,
			name:      "missing order number in request body case",
			body: body{
				Sum: 10,
			},
			want: http.StatusBadRequest,
		},
		{
			repoState: repoState{},
			authUser:  defaultAuthUser,
			name:      "missing sum in request body",
			body: body{
				Order: validLuhnString,
			},
			want: http.StatusBadRequest,
		},
		{
			repoState: repoState{},
			authUser:  defaultAuthUser,
			name:      "not valid luhn order number",
			body: body{
				Order: "123",
				Sum:   50,
			},
			want: http.StatusUnprocessableEntity,
		},
		{
			repoState: repoState{},
			authUser:  defaultAuthUser,
			name:      "not enough bonus on user balance",
			body: body{
				Order: validLuhnString,
				Sum:   200,
			},
			want: http.StatusPaymentRequired,
		},
		{
			repoState: repoState{
				updateUserBalanceReturnErr: true,
			},
			authUser: defaultAuthUser,
			name:     "update user balance return err",
			body: body{
				Order: validLuhnString,
				Sum:   50,
			},
			want: http.StatusInternalServerError,
		},
		{
			repoState: repoState{
				saveWithdrawalReturnErr: true,
			},
			authUser: defaultAuthUser,
			name:     "save withdrawal return err",
			body: body{
				Order: validLuhnString,
				Sum:   50,
			},
			want: http.StatusInternalServerError,
		},
	}
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := context.WithValue(ctx, entity.UserKey, tt.authUser)
			repo := handler_mock.NewMockWithdrawStorage(ctrl)
			newUserBalance := tt.authUser.Balance - tt.body.Sum
			if tt.repoState.updateUserBalanceReturnErr {
				repo.EXPECT().
					UpdateUserBalanceByID(userCtx, tt.authUser.ID, newUserBalance).
					Return(fmt.Errorf("failed update user balance")).
					AnyTimes()
			} else {
				repo.EXPECT().
					UpdateUserBalanceByID(userCtx, tt.authUser.ID, newUserBalance).
					Return(nil).
					AnyTimes()
			}
			intOrderNumber, err := strconv.Atoi(tt.body.Order)
			if err != nil {
				intOrderNumber = 0
			}
			if tt.repoState.saveWithdrawalReturnErr {
				repo.EXPECT().
					SaveWithdrawal(userCtx, entity.Withdrawal{
						OrderNumber: intOrderNumber,
						UserID:      tt.authUser.ID,
						Sum:         tt.body.Sum,
					}).Return(fmt.Errorf("failed save withdrawal")).
					AnyTimes()
			} else {
				repo.EXPECT().
					SaveWithdrawal(userCtx, entity.Withdrawal{
						OrderNumber: intOrderNumber,
						UserID:      tt.authUser.ID,
						Sum:         tt.body.Sum,
					}).Return(nil).
					AnyTimes()
			}
			byteBody, err := json.Marshal(&tt.body)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw",
				bytes.NewReader(byteBody))
			request = request.WithContext(userCtx)
			writer := httptest.NewRecorder()
			handler := WithdrawBonus(repo)
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			_ = res.Body.Close()
			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
