package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"wishlist/internal/api/controllers"
	"wishlist/internal/api/middlewares"
	"wishlist/internal/config"
	"wishlist/internal/logger"
	"wishlist/internal/utils/colors"
	"wishlist/internal/utils/gin"
)

type API struct {
	engine   *gin.Engine
	userCtrl *controllers.UsersController
}

func NewAPI(engine *gin.Engine, uc *controllers.UsersController) *API {
	return &API{engine: engine, userCtrl: uc}
}

func NewEngine() *gin.Engine {
	if viper.GetBool(config.GinReleaseMode) {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	_ = engine.SetTrustedProxies(nil)
	return engine
}

func (api *API) RegisterMiddlewares() {
	api.engine.Use(gin.RecoveryWithWriter(logger.GetWriters()))
	api.engine.Use(middlewares.RequestID())
	api.engine.Use(ginutils.CustomGinLogger(logger.GetWriters()))
	api.engine.Use(middlewares.CORS())
}

func (api *API) RegisterRoutes() {
	api.userCtrl.RegisterRoutes()
}

func (api *API) Run() {
	fmt.Printf("Starting Gin engine...")

	addr := fmt.Sprintf("%s:%s", viper.GetString(config.ApiHost), viper.GetString(config.ApiPort))
	if ln, err := net.Listen("tcp", addr); err != nil {
		fmt.Println()
		logger.Fatalf("Port %s is already in use: %v", viper.GetString(config.ApiPort), err)
	} else {
		_ = ln.Close()
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.engine,
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(exit)

	go func() {
		fmt.Println(colors.Green("    Done."))
		logger.Info("Listening on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Server stopped unexpectedly: %v", err)
		}
	}()

	sig := <-exit
	logger.Info("Received %s signal, shutting down...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Graceful shutdown failed: %v", err)
	}

	logger.Info("Server stopped.")
}
