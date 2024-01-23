package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dto"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

const invalidUserCredentialsMessage = "invalid user credentials"

type LoginStorage interface {
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
}

func Login(loginStorage LoginStorage, gopherConfig *config.GophermartConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := dto.NewUserCredentialsFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		user, err := loginStorage.GetUserByLogin(request.Context(), body.Login)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(writer, invalidUserCredentialsMessage, http.StatusUnauthorized)
				return
			}
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		passValid := user.CheckPassword(body.Password)
		if !passValid {
			http.Error(writer, invalidUserCredentialsMessage, http.StatusUnauthorized)
			return
		}

		expTime := time.Hour * time.Duration(gopherConfig.ExpTime)
		token, err := user.GenerateJWTToken(gopherConfig.SecretKey, expTime)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:    "AUTH_TOKEN",
			Value:   token,
			Expires: time.Now().Add(expTime),
		})
		writer.WriteHeader(http.StatusOK)
		_, err = fmt.Fprint(writer, "you successfully login")
		if err != nil {
			logger.Log.Warnf("failed send success login message to client")
			return
		}
	}
}
