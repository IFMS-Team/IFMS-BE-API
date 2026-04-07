package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/handler"

	_ "IFMS-BE-API/docs"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/IFMS-Team/IFMS-BE/database"
	redisclient "github.com/IFMS-Team/IFMS-BE/redis"
)

// @title           IFMS API
// @version         1.0
// @description     IFMS Backend API Service

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	_ = godotenv.Load()

	logger.Info("IFMS API Service starting...")

	db := database.NewDbPostgres(logger)
	defer db.Close()

	redisclient.Connect()
	defer redisclient.Close()

	env := os.Getenv("APP_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	engine.Use(corsMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	jwtSecret := os.Getenv("JWT_SECRET")

	appCtx := &app.AppContext{
		Engine:     engine,
		DbPostgres: db,
		Pool:       db.Pool,
		Queries:    db.Q,
		Logger:     logger,
		Validator:  validator.New(),
		Redis:      redisclient.GetClient(),
		KeySecret:  []byte(jwtSecret),
	}

	handler.NewAuthHandler(appCtx)
	handler.NewRoleHandler(appCtx)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: engine,
	}

	go func() {
		logger.Info("Server is running", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
