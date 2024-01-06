package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

const TokenExp = time.Hour * 12
const SecretKey = "secret-key"
const failedSendDataErrPattern = "failed send data to client: %v"

type RegisterBody struct {
	Login    string `json:"login,omitempty" validate:"required,min=3"`
	Password string `json:"password,omitempty" validate:"required,min=3"`
}

type RegisterStorage interface {
	HasUser(ctx context.Context, login string) (bool, error)
	SaveUser(ctx context.Context, user entity.User) error
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
}

func Register(repo RegisterStorage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body := RegisterBody{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			http.Error(writer, fmt.Errorf("failed decode register body: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(&body)
		if err != nil {
			http.Error(writer, fmt.Errorf("invalid body given: %w", err).Error(), http.StatusBadRequest)
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

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
			UserID: registeredUser.ID,
		})

		tokenString, err := token.SignedString([]byte(SecretKey))
		if err != nil {
			http.Error(writer, fmt.Errorf("failed signed token: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(writer, &http.Cookie{
			Name:    "AUTH_TOKEN",
			Value:   tokenString,
			Expires: time.Now().Add(TokenExp),
		})

		if _, err = fmt.Fprint(writer, "Welcome my friend. You was successfully registered"); err != nil {
			logger.Log.Warnf(failedSendDataErrPattern, err)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}
