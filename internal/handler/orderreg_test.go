package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ilya372317/gophermart/internal/entity"
	handler_mock "github.com/ilya372317/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterOrder(t *testing.T) {
	defaultAuthUser := &entity.User{
		CreatedAT: time.Now(),
		UpdatedAT: time.Now(),
		Login:     "test",
		Password:  "123",
		ID:        1,
	}
	validLunNumberString := "12345678903"
	type data struct {
		saveOrderReturnErr                 bool
		hasOrderByNumberReturnErr          bool
		hasOrderByNumberAndUserIDReturnErr bool
		hasOrderByNumber                   bool
		hasOrderByNumberAndUserID          bool
	}
	tests := []struct {
		name     string
		data     data
		argument string
		authUser *entity.User
		want     int
	}{
		{
			name:     "wrong body format case",
			data:     data{},
			argument: "wron-format-expected-int",
			authUser: defaultAuthUser,
			want:     http.StatusBadRequest,
		},
		{
			name:     "empty body case",
			data:     data{},
			argument: "",
			authUser: defaultAuthUser,
			want:     http.StatusBadRequest,
		},
		{
			name:     "invalid lun number case",
			data:     data{},
			argument: "1234567890",
			authUser: defaultAuthUser,
			want:     http.StatusUnprocessableEntity,
		},
		{
			name: "has order by number and user id return err case",
			data: data{
				hasOrderByNumberAndUserIDReturnErr: true,
				hasOrderByNumberAndUserID:          false,
			},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusInternalServerError,
		},
		{
			name: "user already registered order case",
			data: data{
				hasOrderByNumberAndUserIDReturnErr: false,
				hasOrderByNumberAndUserID:          true,
			},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusOK,
		},
		{
			name: "has order by number return err case",
			data: data{
				hasOrderByNumberReturnErr: true,
				hasOrderByNumber:          false,
			},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusInternalServerError,
		},
		{
			name: "other user registered this order case",
			data: data{
				hasOrderByNumberReturnErr: false,
				hasOrderByNumber:          true,
			},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusConflict,
		},
		{
			name: "save order return err case",
			data: data{
				saveOrderReturnErr: true,
			},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusInternalServerError,
		},
		{
			name:     "success register order case",
			data:     data{},
			argument: validLunNumberString,
			authUser: defaultAuthUser,
			want:     http.StatusAccepted,
		},
	}
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestContext := context.WithValue(ctx, entity.UserKey, tt.authUser)
			repo := handler_mock.NewMockRegisterOrderStorage(ctrl)
			intArgument, parseIntErr := strconv.Atoi(tt.argument)
			newOrder := &entity.Order{
				Number: intArgument,
				UserID: tt.authUser.ID,
			}
			if parseIntErr == nil {
				if tt.data.hasOrderByNumberReturnErr {
					repo.EXPECT().HasOrderByNumber(requestContext, gomock.Eq(intArgument)).
						Return(false, fmt.Errorf("failed check has order by number")).
						AnyTimes()
				} else {
					repo.EXPECT().HasOrderByNumber(requestContext, gomock.Eq(intArgument)).
						Return(tt.data.hasOrderByNumber, nil).
						AnyTimes()
				}
				if tt.data.saveOrderReturnErr {
					repo.EXPECT().SaveOrder(requestContext, newOrder).
						Return(fmt.Errorf("failed save order")).
						AnyTimes()
				} else {
					repo.EXPECT().SaveOrder(requestContext, newOrder).
						Return(nil).
						AnyTimes()
				}
				if tt.data.hasOrderByNumberAndUserIDReturnErr {
					repo.EXPECT().HasOrderByNumberAndUserID(requestContext, intArgument, tt.authUser.ID).
						Return(false, fmt.Errorf("failed check order by number and user id")).
						AnyTimes()
				} else {
					repo.EXPECT().HasOrderByNumberAndUserID(requestContext, intArgument, tt.authUser.ID).
						Return(tt.data.hasOrderByNumberAndUserID, nil).
						AnyTimes()
				}
			}
			orderProcessor := handler_mock.NewMockOrderProcessor(ctrl)
			orderProcessor.EXPECT().ProcessOrder(intArgument).AnyTimes()
			request := httptest.NewRequest(http.MethodPost,
				"/api/user/orders", bytes.NewReader([]byte(tt.argument)))
			request = request.WithContext(requestContext)
			writer := httptest.NewRecorder()
			handler := RegisterOrder(repo)
			handler.ServeHTTP(writer, request)

			res := writer.Result()
			err := res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}
