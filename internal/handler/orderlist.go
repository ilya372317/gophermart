package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilya372317/gophermart/internal/entity"
)

type OrderListStorage interface {
	GetOrderListByUserID(context.Context, uint) ([]entity.Order, error)
}

func GetOrderList(repo OrderListStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authUser, ok := request.Context().Value(entity.UserKey).(*entity.User)
		if !ok {
			http.Error(writer, "unauthorized", http.StatusInternalServerError)
			return
		}
		orders, err := repo.GetOrderListByUserID(request.Context(), authUser.ID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(writer, "orders: %v", orders)
	}
}
