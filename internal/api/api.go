package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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
	api.engine.Use(gin.RecoveryWithWriter(io.MultiWriter(os.Stdout, logger.GetLogFile())))
	api.engine.Use(middlewares.RequestID())
	api.engine.Use(logger.CustomGinLogger(io.MultiWriter(os.Stdout, logger.GetLogFile())))
	api.engine.Use(middlewares.CORS())
}

func (api *API) RegisterRoutes() {
	api.userCtrl.RegisterRoutes()
}

func (api *API) Run() {
	addr := fmt.Sprintf(":%s", viper.GetString(config.ApiPort))

	if ln, err := net.Listen("tcp", addr); err != nil {
		log.Fatalf("Port %s is already in use: %v", viper.GetString(config.ApiPort), err)
	} else {
		_ = ln.Close()
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: api.engine,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("API server listening on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped.")
}
