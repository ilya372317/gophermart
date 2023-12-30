package router

import "github.com/go-chi/chi/v5"

func New() *chi.Mux {
	r := chi.NewRouter()

	return r
}
