package orderproc

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/accrual"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/entity"
	orderproc_mock "github.com/ilya372317/gophermart/internal/orderproc/mocks"
	"github.com/stretchr/testify/require"
)

func TestOrderProcessor_UpdateStatusToProcessingReturnErr(t *testing.T) {
	argument := 123
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := orderproc_mock.NewMockAccrualClient(ctrl)
	strg := orderproc_mock.NewMockOrderStorage(ctrl)
	strg.EXPECT().UpdateOrderStatusByNumber(ctx, argument, entity.StatusProcessing).
		Return(fmt.Errorf("failed update status to PROCESSING")).MinTimes(1).MaxTimes(1)
	o := New(client, strg)
	err := o.processOrder(ctx, &config.GophermartConfig{}, argument)
	require.Error(t, err)
}

func TestOrderProcessor_SendRequestToAccrualReturnErr(t *testing.T) {
	argument := 123
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := orderproc_mock.NewMockAccrualClient(ctrl)
	client.EXPECT().GetCalculation(argument).
		Return(nil, fmt.Errorf("failed send request to accrual")).
		MinTimes(1).
		MaxTimes(1)
	strg := orderproc_mock.NewMockOrderStorage(ctrl)
	strg.EXPECT().UpdateOrderStatusByNumber(ctx, argument, entity.StatusProcessing).
		Return(nil).MinTimes(1).MaxTimes(1)
	strg.EXPECT().UpdateOrderStatusByNumber(ctx, argument, entity.StatusInvalid).
		Return(nil).
		MinTimes(1).
		MaxTimes(1)
	o := New(client, strg)
	err := o.processOrder(ctx, &config.GophermartConfig{}, argument)
	require.Error(t, err)
}

func TestOrderProcessor_ProcessOrderSuccessCase(t *testing.T) {
	const userID = uint(1)
	const userBalance = 10
	const accrualResponseAmount = float64(10)
	const orderID = 1
	const bonusToSet = float64(20)
	argument := 123
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := orderproc_mock.NewMockAccrualClient(ctrl)
	client.EXPECT().GetCalculation(argument).
		Return(&accrual.CalculationResponse{
			Order:      "123",
			Status:     entity.StatusProcessed,
			Accrual:    accrualResponseAmount,
			StatusCode: http.StatusOK,
		}, nil).
		MinTimes(1).
		MaxTimes(1)
	strg := orderproc_mock.NewMockOrderStorage(ctrl)
	strg.EXPECT().UpdateOrderStatusByNumber(ctx, argument, entity.StatusProcessing).
		Return(nil).MinTimes(1).MaxTimes(1)
	strg.EXPECT().GetOrderByNumber(ctx, argument).
		Return(&entity.Order{
			CreatedAT: time.Now(),
			UpdatedAT: time.Now(),
			Status:    "PROCESSING",
			Number:    argument,
			Accrual: sql.NullFloat64{
				Float64: 0,
				Valid:   false,
			},
			ID:     orderID,
			UserID: userID,
		}, nil).MinTimes(1)
	strg.EXPECT().GetUserByID(ctx, userID).Return(&entity.User{
		CreatedAT: time.Now(),
		UpdatedAT: time.Now(),
		Login:     "123",
		Password:  "123",
		ID:        userID,
		Balance:   userBalance,
	}, nil).MinTimes(1)
	strg.EXPECT().UpdateUserBalanceByID(ctx, userID, bonusToSet).Return(nil).MinTimes(1)
	strg.EXPECT().UpdateOrderAccrualByNumber(ctx, argument, sql.NullFloat64{
		Float64: accrualResponseAmount,
		Valid:   true,
	}).Return(nil).MinTimes(1)
	strg.EXPECT().UpdateOrderStatusByNumber(ctx, argument, entity.StatusProcessed).Return(nil).MinTimes(1)
	o := New(client, strg)
	err := o.processOrder(ctx, &config.GophermartConfig{}, argument)
	require.NoError(t, err)
}
