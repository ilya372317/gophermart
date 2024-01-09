package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dto"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

const failedSendDataErrPattern = "failed send data to client: %v"

type RegisterStorage interface {
	HasUser(ctx context.Context, login string) (bool, error)
	SaveUser(ctx context.Context, user entity.User) error
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
}

func Register(repo RegisterStorage, gopherConfig *config.GophermartConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := dto.NewUserCredentialsFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		userAlreadyRegistered, err := repo.HasUser(request.Context(), body.Login)
		if err != nil {
			http.Error(writer, fmt.Errorf("falied check user registration: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		if userAlreadyRegistered {
			http.Error(writer, "login is taken by another user", http.StatusConflict)
			return
		}

		user := entity.User{
			Login: body.Login,
		}
		user.SetPassword(body.Password)
		saveErr := repo.SaveUser(request.Context(), user)
		if saveErr != nil {
			http.Error(writer, fmt.Errorf("failed save user to database: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		registeredUser, err := repo.GetUserByLogin(request.Context(), body.Login)
		if err != nil {
			http.Error(writer, fmt.Errorf("failed get registered user: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		tokenExp := time.Hour * time.Duration(gopherConfig.ExpTime)
		tokenString, err := registeredUser.GenerateJWTToken(gopherConfig.SecretKey, tokenExp)
		if err != nil {
			http.Error(writer, fmt.Errorf("failed signed token: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:    "AUTH_TOKEN",
			Value:   tokenString,
			Expires: time.Now().Add(tokenExp),
		})

		if _, err = fmt.Fprint(writer, "Welcome my friend. You was successfully registered"); err != nil {
			logger.Log.Warnf(failedSendDataErrPattern, err)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}
