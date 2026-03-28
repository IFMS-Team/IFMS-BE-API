package api

import (
	"context"
	"net/http"
	"os"
	"time"

	"IFMS-BE-API/internal/router"

	"github.com/vippergod12/IFMS-BE/database"
	redisclient "github.com/vippergod12/IFMS-BE/redis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Application struct {
	ctx    context.Context
	logger *zap.Logger
	db     *database.DbPostgres
	server *http.Server
}

func NewApiApplication(ctx context.Context) *Application {
	logger, _ := zap.NewProduction()

	_ = godotenv.Load()

	logger.Info("IFMS API Service starting...")

	db := database.NewDbPostgres(logger)

	if err := db.AutoMigrate("../IFMS-BE/sql/schema"); err != nil {
		logger.Fatal("Migration failed", zap.Error(err))
	}

	redisclient.Connect()

	env := os.Getenv("APP_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.Setup(db, logger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	return &Application{
		ctx:    ctx,
		logger: logger,
		db:     db,
		server: server,
	}
}

func (a *Application) Start() {
	go func() {
		a.logger.Info("Server is running", zap.String("addr", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Server failed", zap.Error(err))
		}
	}()
}

func (a *Application) Shutdown() {
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", zap.Error(err))
	}

	redisclient.Close()
	a.db.Close()
	a.logger.Sync()

	a.logger.Info("Server stopped")
}
