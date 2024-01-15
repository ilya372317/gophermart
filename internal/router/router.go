package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/gmiddleware"
	"github.com/ilya372317/gophermart/internal/handler"
)

type MarketStorage interface {
	HasUser(ctx context.Context, login string) (bool, error)
	SaveUser(ctx context.Context, user entity.User) error
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
}

func New(repo MarketStorage, gopherConfig *config.GophermartConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Group(func(r chi.Router) {
		r.Use(gmiddleware.ShouldHasBody)
		r.Post("/register", handler.Register(repo, gopherConfig))
		r.Post("/login", handler.Login(repo, gopherConfig))
	})
	r.Group(func(r chi.Router) {
		r.Use(gmiddleware.Auth(gopherConfig, repo))
		r.Get("/test", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprint(writer, "test is work")
		})
	})
	return r
}
