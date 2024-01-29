package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/gmiddleware"
	"github.com/ilya372317/gophermart/internal/handler"
	"github.com/ilya372317/gophermart/internal/storage"
)

func New(
	repo *storage.DBStorage,
	gopherConfig *config.GophermartConfig,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gmiddleware.JSONResponse())
	r.Use(gmiddleware.Compressed())
	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(gmiddleware.ShouldHasBody)
			r.Post("/register", handler.Register(repo, gopherConfig))
			r.Post("/login", handler.Login(repo, gopherConfig))
		})
		r.Group(func(r chi.Router) {
			r.Use(gmiddleware.Auth(gopherConfig, repo))
			r.Post("/user/balance/withdraw", handler.WithdrawBonus(repo))
			r.Post("/user/orders", handler.RegisterOrder(repo))
			r.Get("/user/orders", handler.GetOrderList(repo))
			r.Get("/user/withdrawals", handler.WithdrawalList(repo))
			r.Get("/user/balance", handler.GetUserBalance(repo))
		})
	})

	return r
}
