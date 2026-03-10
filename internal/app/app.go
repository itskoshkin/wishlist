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
	"wishlist/pkg/minio"
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

	// Databases/clients
	db, err := postgres.NewInstance(ctx, config.PostgresConfig())
	if err != nil {
		logger.Fatal(err)
	}
	rc, err := redis.NewClient(ctx, config.RedisConfig())
	if err != nil {
		logger.Fatal(err)
	}
	s3, err := minio.NewClient(ctx, config.MinioConfig())
	if err != nil {
		logger.Fatal(err)
	}

	// Storages
	userStore := storage.NewUserStorage(db)
	wishStore := storage.NewWishStorage(db)
	listStore := storage.NewListStorage(db)
	tokenStore := storage.NewTokenStorage(rc)

	// Services
	authSvc := services.NewAuthService(tokenStore)
	emailSvc := services.NewEmailService()
	minioSvc := storage.NewMinioService(s3)
	userSvc := services.NewUserService(emailSvc, userStore, tokenStore, minioSvc, logger.GlobalLogger{})
	listSvc := services.NewListService(listStore, wishStore)
	wishSvc := services.NewWishService(wishStore, listStore)

	// API
	e := api.NewEngine()
	mw := middlewares.NewMiddlewares(authSvc)

	// Controllers
	webCtrl := controllers.NewWebController()
	userCtrl := controllers.NewUsersController(e, mw, authSvc, userSvc)
	listCtrl := controllers.NewListsController(e, mw, listSvc)
	wishCtrl := controllers.NewWishesController(e, mw, wishSvc)

	return &App{API: api.NewAPI(e, webCtrl, userCtrl, listCtrl, wishCtrl)}
}

func (a *App) Run() {
	a.API.RegisterMiddlewares()
	a.API.RegisterRoutes()
	a.API.Run()
}
