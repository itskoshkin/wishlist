package app

import (
	"context"
	"time"

	"wishlist/internal/api"
	"wishlist/internal/api/controllers"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/logger"
	"wishlist/internal/services"
	"wishlist/internal/storage"
	"wishlist/pkg/postgres"
	"wishlist/pkg/redis"
)

type App struct {
	API *api.API
}

func Load() *App {
	config.LoadConfig()
	logger.SetupLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := postgres.NewInstance(ctx, config.PostgresConfig())
	if err != nil {
		logger.Fatal(err)
	}

	rc, err := redis.NewClient(ctx, config.RedisConfig())
	if err != nil {
		logger.Fatal(err)
	}

	e := api.NewEngine()
	st := storage.NewUserStorage(db)
	ts := storage.NewTokenStorage(rc)
	as := services.NewAuthService(ts)
	es := services.NewEmailService()
	us := services.NewUserService(es, st, ts, logger.GlobalLogger{})
	mw := middlewares.NewMiddlewares(as)
	uc := controllers.NewUsersController(e, mw, as, us)

	return &App{API: api.NewAPI(e, uc)}
}

func (a *App) Run() {
	a.API.RegisterMiddlewares()
	a.API.RegisterRoutes()
	a.API.Run()
}
