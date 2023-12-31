package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilya372317/gophermart/internal/handler"
)

func New() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", handler.Register())
	r.Post("/login", handler.Login())
	return r
}
