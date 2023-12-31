package handler

import (
	"fmt"
	"net/http"
)

func Login() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprint(writer, "Here we will have register route")
		if err != nil {
			http.Error(writer, fmt.Errorf("failed write data").Error(), http.StatusInternalServerError)
		}
	}
}
