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
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"wishlist/docs"
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
	listCtrl *controllers.ListsController
	wishCtrl *controllers.WishesController
}

func NewAPI(e *gin.Engine, uc *controllers.UsersController, lc *controllers.ListsController, wc *controllers.WishesController) *API {
	return &API{
		engine:   e,
		userCtrl: uc,
		listCtrl: lc,
		wishCtrl: wc,
	}
}

func NewEngine() *gin.Engine {
	if viper.GetBool(config.GinReleaseMode) {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	_ = engine.SetTrustedProxies(nil) // Can nil produce an error? Or can a robot write a symphony?
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
	api.listCtrl.RegisterRoutes()
	api.wishCtrl.RegisterRoutes()
	// Swagger
	{
		{
			docs.SwaggerInfo.Host = fmt.Sprintf("%s", viper.GetString(config.WebAppDomain))
			docs.SwaggerInfo.BasePath = viper.GetString(config.ApiBasePath)
		}
		api.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
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
