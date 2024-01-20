package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilya372317/gophermart/internal/dto"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

type RegisterOrderStorage interface {
	HasOrderByNumber(context.Context, int) (bool, error)
	HasOrderByNumberAndUserID(context.Context, int, uint) (bool, error)
	SaveOrder(context.Context, *entity.Order) error
}

func RegisterOrder(gopherStorage RegisterOrderStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		registerOrder, err := dto.CreateRegisterOrderFormRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if !isValidLun(registerOrder.Number) {
			http.Error(writer, "invalid order number", http.StatusUnprocessableEntity)
			return
		}
		authUser, exists := request.Context().Value(entity.UserKey).(*entity.User)
		if !exists {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
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
		//TODO: 4.send queue job on processingsd
		go func(number int) {

		}(newOrder.Number)
		writer.WriteHeader(http.StatusAccepted)
		if _, err := fmt.Fprint(writer, "order registered"); err != nil {
			logger.Log.Warnf("faield write response: %v", err)
		}
	}
}

func isValidLun(number int) bool {
	const baseNumber = 10
	return (number%baseNumber+checksum(number/baseNumber))%baseNumber == 0
}

func checksum(number int) int {
	var luhn int
	const ten, two, zero, nine = 10, 2, 0, 9

	for i := zero; number > zero; i++ {
		cur := number % ten

		if i%two == zero {
			cur *= two
			if cur > nine {
				cur = cur%ten + cur/ten
			}
		}

		luhn += cur
		number /= ten
	}
	return luhn % ten
}
