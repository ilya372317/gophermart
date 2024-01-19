package dto

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type RegisterOrder struct {
	Number int `validate:"required"`
}

func CreateRegisterOrderFormRequest(r *http.Request) (*RegisterOrder, error) {
	dto := &RegisterOrder{}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read register order body: %w", err)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("request body should not be empty")
	}
	intNumber, err := strconv.Atoi(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed parse order number to int: %w", err)
	}
	dto.Number = intNumber
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(dto); err != nil {
		return nil, fmt.Errorf("body is invalid: %w", err)
	}

	return dto, nil
}
