package router

import (
	"net/http"

	"IFMS-BE-API/internal/handler"
	"IFMS-BE-API/internal/repo"
	"IFMS-BE-API/internal/service"
	"IFMS-be/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Timezone")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Setup(db *database.DbPostgres, logger *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repo.NewUserRepo(db.Q)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, logger)

	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.POST("", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}

	return r
}
