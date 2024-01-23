package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/ilya372317/gophermart/internal/luhnalgo"
)

type RegisterOrderStorage interface {
	HasOrderByNumber(context.Context, int) (bool, error)
	HasOrderByNumberAndUserID(context.Context, int, uint) (bool, error)
	SaveOrder(context.Context, *entity.Order) error
	UpdateOrderStatusByNumber(context.Context, int, string) error
	GetOrderByNumber(context.Context, int) (*entity.Order, error)
	UpdateUserBalanceByID(ctx context.Context, id uint, balance int) error
}

type OrderProcessor interface {
	ProcessOrder(int)
}

type RegisterOrderRequest struct {
	Number int `validate:"required"`
}

func CreateRegisterOrderFormRequest(r *http.Request) (*RegisterOrderRequest, error) {
	dto := &RegisterOrderRequest{}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read register order body: %w", err)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("request body should not be empty")
	}
	intNumber, err := strconv.Atoi(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed parse order number to int: %w", err)
	}
	dto.Number = intNumber
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(dto); err != nil {
		return nil, fmt.Errorf("body is invalid: %w", err)
	}

	return dto, nil
}

func RegisterOrder(
	gopherStorage RegisterOrderStorage,
	orderProcessor OrderProcessor,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		registerOrder, err := CreateRegisterOrderFormRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if !luhnalgo.IsValid(registerOrder.Number) {
			http.Error(writer, "invalid order number", http.StatusUnprocessableEntity)
			return
		}
		authUser, exists := request.Context().Value(entity.UserKey).(*entity.User)
		if !exists {
			http.Error(writer, errUnauthorized, http.StatusUnauthorized)
			return
		}

		authUserHasOrder, err := gopherStorage.HasOrderByNumberAndUserID(request.Context(),
			registerOrder.Number, authUser.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if authUserHasOrder {
			writer.WriteHeader(http.StatusOK)
			if _, err = fmt.Fprintf(writer, "user already registered order"); err != nil {
				logger.Log.Warnf("failed send response about registered order")
			}
			return
		}

		orderExists, err := gopherStorage.HasOrderByNumber(request.Context(), registerOrder.Number)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if orderExists {
			http.Error(writer, "order registered by other user", http.StatusConflict)
			return
		}

		newOrder := &entity.Order{
			Number: registerOrder.Number,
			UserID: authUser.ID,
		}
		if err = gopherStorage.SaveOrder(request.Context(), newOrder); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		go orderProcessor.ProcessOrder(newOrder.Number)
		writer.WriteHeader(http.StatusAccepted)
		if _, err := fmt.Fprint(writer, "order registered"); err != nil {
			logger.Log.Warnf("faield write response: %v", err)
		}
	}
}
