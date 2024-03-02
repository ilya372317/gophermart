package dto

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type UserCredentials struct {
	Login    string `json:"login,omitempty" validate:"required,min=3"`
	Password string `json:"password,omitempty" validate:"required,min=3"`
}

func NewUserCredentialsFromRequest(request *http.Request) (*UserCredentials, error) {
	dto := &UserCredentials{}
	err := json.NewDecoder(request.Body).Decode(dto)
	if err != nil {
		return nil, fmt.Errorf("failed decode user credentials: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(dto)
	if err != nil {
		return nil, fmt.Errorf("invalid data given: %w", err)
	}

	return dto, nil
}
