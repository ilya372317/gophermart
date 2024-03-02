package accrual

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type CalculationResponse struct {
	Order      string  `json:"order"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	StatusCode int
}

type Client struct {
	c *resty.Client
}

func New(accrualHost string) *Client {
	return &Client{
		c: resty.New().SetBaseURL(accrualHost),
	}
}

func (a *Client) GetCalculation(number int) (*CalculationResponse, error) {
	response := &CalculationResponse{}
	stringNumber := strconv.FormatInt(int64(number), 10)
	res, err := a.c.R().SetResult(response).
		Get("/api/orders/" + stringNumber)
	if err != nil {
		return nil, fmt.Errorf("failed make request to accrual for calculation: %w", err)
	}
	response.StatusCode = res.StatusCode()
	return response, nil
}
