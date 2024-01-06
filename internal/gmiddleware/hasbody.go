package gmiddleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func ShouldHasBody(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		bodyContent, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer,
				fmt.Errorf("failed read body for check len: %w", err).Error(),
				http.StatusInternalServerError)
			return
		}
		if len(bodyContent) == 0 {
			http.Error(writer, "request should have body", http.StatusBadRequest)
			return
		}
		request.Body = io.NopCloser(bytes.NewReader(bodyContent))
		h.ServeHTTP(writer, request)
	})
}
