package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

type WithdrawalResponse struct {
	ProcessedAT time.Time `json:"processed_at"`
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
}

func transformWithdrawalsToResponse(withdrawals []entity.Withdrawal) []WithdrawalResponse {
	responseList := make([]WithdrawalResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		responseItem := WithdrawalResponse{
			ProcessedAT: withdrawal.CreatedAT,
			Order:       strconv.Itoa(withdrawal.OrderNumber),
			Sum:         withdrawal.Sum,
		}
		responseList = append(responseList, responseItem)
	}

	return responseList
}

type WithdrawalListStorage interface {
	GetWithdrawalListByUserID(context.Context, uint) ([]entity.Withdrawal, error)
}

func WithdrawalList(repo WithdrawalListStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authUser, ok := request.Context().Value(entity.UserKey).(*entity.User)
		if !ok {
			http.Error(writer, errUnauthorized, http.StatusInternalServerError)
			return
		}
		withdrawals, err := repo.GetWithdrawalListByUserID(request.Context(), authUser.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(withdrawals) == 0 {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		responseList := transformWithdrawalsToResponse(withdrawals)
		responseString, err := json.Marshal(responseList)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err = fmt.Fprint(writer, string(responseString)); err != nil {
			logger.Log.Error(err.Error())
		}
	}
}
