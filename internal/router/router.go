package router

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/handler"
)

type MarketStorage interface {
	HasUser(ctx context.Context, login string) (bool, error)
	SaveUser(ctx context.Context, user entity.User) error
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
}

func New(repo MarketStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/register", handler.Register(repo))
	r.Post("/login", handler.Login())
	r.Group(func(r chi.Router) {

	})
	return r
}
