package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/luhnalgo"
)

type WithdrawRequest struct {
	Order string `json:"order" valid:"required,numeric"`
	Sum   int    `json:"sum" valid:"required"`
}

func createWithdrawFromRequest(r *http.Request) (WithdrawRequest, error) {
	dto := WithdrawRequest{}
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return WithdrawRequest{}, fmt.Errorf("failed decode witdraw request body: %w", err)
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(dto); err != nil {
		return WithdrawRequest{}, fmt.Errorf("invalid request body: %w", err)
	}

	return dto, nil
}

type WithdrawStorage interface {
	UpdateUserBalanceByID(context.Context, uint, int) error
	SaveWithdrawal(context.Context, entity.Withdrawal) error
}

func WithdrawBonus(repo WithdrawStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		withdrawalRequest, err := createWithdrawFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		intOrderNumber, err := strconv.Atoi(withdrawalRequest.Order)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if !luhnalgo.IsValid(intOrderNumber) {
			http.Error(writer, "invalid order number", http.StatusUnprocessableEntity)
			return
		}

		authUser, ok := request.Context().Value(entity.UserKey).(*entity.User)
		if !ok {
			http.Error(writer, "unauthorized", http.StatusInternalServerError)
			return
		}
		if authUser.Balance < withdrawalRequest.Sum {
			http.Error(writer, "there are insufficient funds in the account", http.StatusPaymentRequired)
			return
		}

		newBalance := authUser.Balance - withdrawalRequest.Sum
		if err = repo.UpdateUserBalanceByID(request.Context(), authUser.ID, newBalance); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		withdrawal := entity.Withdrawal{
			OrderNumber: intOrderNumber,
			UserID:      authUser.ID,
			Sum:         withdrawalRequest.Sum,
		}
		if err = repo.SaveWithdrawal(request.Context(), withdrawal); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}
