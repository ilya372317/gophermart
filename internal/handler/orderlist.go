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

type OrderListStorage interface {
	GetOrderListByUserID(context.Context, uint) ([]entity.Order, error)
}

type OrderResponse struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int64     `json:"accrual,omitempty"`
}

func TransformOrderListToOrderResponseList(orderList []entity.Order) []OrderResponse {
	responseList := make([]OrderResponse, 0, len(orderList))
	for _, order := range orderList {
		response := OrderResponse{
			Number:     strconv.Itoa(order.Number),
			Status:     order.Status,
			UploadedAt: order.CreatedAT,
		}
		if order.Accrual.Valid {
			response.Accrual = order.Accrual.Int64
		}
		responseList = append(responseList, response)
	}

	return responseList
}

func GetOrderList(repo OrderListStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authUser, ok := request.Context().Value(entity.UserKey).(*entity.User)
		if !ok {
			http.Error(writer, errUnauthorized, http.StatusInternalServerError)
			return
		}
		orders, err := repo.GetOrderListByUserID(request.Context(), authUser.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(orders) == 0 {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		responseList := TransformOrderListToOrderResponseList(orders)
		response, err := json.Marshal(responseList)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		if _, err = fmt.Fprint(writer, string(response)); err != nil {
			logger.Log.Warnf("failed send order data: %v", err)
		}
	}
}
