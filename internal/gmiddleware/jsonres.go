package gmiddleware

import "net/http"

func JSONResponse() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			request.Header.Set("Content-Type", "application/json")
			h.ServeHTTP(writer, request)
		})
	}
}
