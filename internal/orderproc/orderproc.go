package orderproc

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ilya372317/gophermart/internal/accrual"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

const setInvalidStatusErrPattern = "failed set invalid status to order: %w"

type OrderStorage interface {
	UpdateOrderStatusByNumber(context.Context, int, string) error
	GetOrderByNumber(context.Context, int) (*entity.Order, error)
	UpdateUserBalanceByID(context.Context, uint, float64) error
	GetUserByID(context.Context, uint) (*entity.User, error)
	UpdateOrderAccrualByNumber(context.Context, int, sql.NullFloat64) error
	GetOrderListByStatus(ctx context.Context, status string) ([]entity.Order, error)
}

type AccrualClient interface {
	GetCalculation(int) (*accrual.CalculationResponse, error)
}

type OrderProcessor struct {
	Repo   OrderStorage
	Client AccrualClient
	TaskCh chan func()
}

func New(client AccrualClient, repo OrderStorage) *OrderProcessor {
	processor := &OrderProcessor{
		Repo:   repo,
		Client: client,
		TaskCh: make(chan func()),
	}
	return processor
}

func (o *OrderProcessor) Start(gopherConfig *config.GophermartConfig) {
	for i := 0; i < gopherConfig.MaxOrderInProcessing; i++ {
		go func() {
			for {
				task, ok := <-o.TaskCh
				if !ok {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Log.Errorf("panic in order processor: %v", r)
						}
					}()
					task()
				}()
			}
		}()
	}
}

func (o *OrderProcessor) processOrder(ctx context.Context, gopherConfig *config.GophermartConfig, number int) error {
	err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusProcessing)
	if err != nil {
		return fmt.Errorf("failed update order status to PROCESSING: %w", err)
	}

	attempts := 0
	var calculationResponse *accrual.CalculationResponse
	for {
		calculationResponse, err = o.Client.GetCalculation(number)
		if err != nil {
			err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid)
			if err != nil {
				return fmt.Errorf(setInvalidStatusErrPattern, err)
			}
			return fmt.Errorf("failed send request to accrual: %w", err)
		}
		if calculationResponse.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Second * time.Duration(gopherConfig.DelayBetweenRequestsToAccrual))
			continue
		}
		if calculationResponse.StatusCode == http.StatusNoContent ||
			calculationResponse.StatusCode == http.StatusInternalServerError {
			if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
				return fmt.Errorf(setInvalidStatusErrPattern, err)
			}
			return fmt.Errorf("order not registered in accrual system")
		}

		if calculationResponse.Status == entity.StatusInvalid {
			if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
				return fmt.Errorf(setInvalidStatusErrPattern, err)
			}
			return fmt.Errorf("failed calculate accrual for order")
		}

		if calculationResponse.Status == entity.StatusProcessed {
			break
		}
		attempts++
		if attempts >= gopherConfig.MaxAccrualRequestAttempts {
			if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
				return fmt.Errorf(setInvalidStatusErrPattern, err)
			}
			return fmt.Errorf("to many attemps to get result from ")
		}
		time.Sleep(time.Second * time.Duration(gopherConfig.DelayBetweenRequestsToAccrual))
	}
	order, err := o.Repo.GetOrderByNumber(ctx, number)
	if err != nil {
		return fmt.Errorf("failed get order for precessing: %w", err)
	}

	user, err := o.Repo.GetUserByID(ctx, order.UserID)
	if err != nil {
		if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
			return fmt.Errorf(setInvalidStatusErrPattern, err)
		}
		return fmt.Errorf("failed get user from order for update balance: %w", err)
	}
	balanceToSet := user.Balance + calculationResponse.Accrual
	if err := o.Repo.UpdateUserBalanceByID(ctx, user.ID, balanceToSet); err != nil {
		if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
			return fmt.Errorf(setInvalidStatusErrPattern, err)
		}
		return fmt.Errorf("failed update user balance: %w", err)
	}

	if calculationResponse.Accrual > 0 {
		if err := o.Repo.UpdateOrderAccrualByNumber(ctx, number, sql.NullFloat64{
			Float64: calculationResponse.Accrual,
			Valid:   true,
		}); err != nil {
			if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusInvalid); err != nil {
				return fmt.Errorf(setInvalidStatusErrPattern, err)
			}
			return fmt.Errorf("failed update order accrual: %w", err)
		}
	}
	if err := o.Repo.UpdateOrderStatusByNumber(ctx, number, entity.StatusProcessed); err != nil {
		logger.Log.Fatalf("failed set PROCESSED status: %v", err)
	}

	return nil
}

func (o *OrderProcessor) registerProcessOrderTask(
	ctx context.Context,
	gopherConfig *config.GophermartConfig,
	number int,
) {
	f := func() {
		if err := o.processOrder(ctx, gopherConfig, number); err != nil {
			logger.Log.Infof("failed process order [%d]: %v", number, err)
			return
		}
	}
	o.TaskCh <- f
}

func (o *OrderProcessor) SupervisingOrders(ctx context.Context, gopherConfig *config.GophermartConfig) {
	timer := time.NewTicker(time.Second)
	for range timer.C {
		select {
		case <-ctx.Done():
			return
		default:
			if err := o.supervisingOrders(ctx, gopherConfig); err != nil {
				logger.Log.Errorf("failed get orders for processing: %v", err)
			}
		}
	}
}

func (o *OrderProcessor) supervisingOrders(ctx context.Context, gopherConfig *config.GophermartConfig) error {
	orderList, err := o.Repo.GetOrderListByStatus(ctx, entity.StatusNew)
	if err != nil {
		return fmt.Errorf("failed get processed orders: %w", err)
	}
	for _, order := range orderList {
		o.registerProcessOrderTask(ctx, gopherConfig, order.Number)
	}

	return nil
}
