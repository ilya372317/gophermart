package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

type UserBalanceStorage interface {
	GetWithdrawalSumByUserID(context.Context, uint) (int, error)
}

type UserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn int     `json:"withdrawn"`
}

func GetUserBalance(repo UserBalanceStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authUser, ok := request.Context().Value(entity.UserKey).(*entity.User)
		if !ok {
			http.Error(writer, "failed get auth user from request", http.StatusInternalServerError)
			return
		}
		withdrawalSum, err := repo.GetWithdrawalSumByUserID(request.Context(), authUser.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		response := UserBalanceResponse{
			Current:   float64(authUser.Balance),
			Withdrawn: withdrawalSum,
		}
		responseContent, err := json.Marshal(&response)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprint(writer, string(responseContent)); err != nil {
			logger.Log.Error(err.Error())
			return
		}
	}
}
