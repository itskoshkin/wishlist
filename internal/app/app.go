package app

import (
	"context"
	"log"
	"time"

	"wishlist/internal/api"
	"wishlist/internal/api/controllers"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/logger"
	"wishlist/internal/services"
	"wishlist/internal/storage"
	"wishlist/pkg/postgres"
)

type App struct {
	API *api.API
}

func Load() *App {
	config.LoadConfig()
	logger.SetupLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := postgres.NewInstance(ctx, config.DatabaseConfig())
	if err != nil {
		log.Fatal(err)
	}

	e := api.NewEngine()
	st := storage.NewUserStorage(db)
	as := services.NewAuthService()
	us := services.NewUserService(st)
	mw := middlewares.NewMiddlewares(as)
	uc := controllers.NewUsersController(e, as, us, mw)
	return &App{API: api.NewAPI(e, uc)}
}

func (a *App) Run() {
	a.API.Run()
}
