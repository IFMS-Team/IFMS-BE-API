package handler

import (
	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/middleware"
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

type AuthHandler struct {
	queries     *db.Queries
	keySecret   []byte
	logger      *zap.Logger
	userService *services.UserService
}

func NewAuthHandler(ctx *app.AppContext) {
	userService := services.NewUserService(ctx.Pool, ctx.Queries, ctx.Logger, nil)

	h := &AuthHandler{
		queries:     ctx.Queries,
		keySecret:   ctx.KeySecret,
		logger:      ctx.Logger,
		userService: userService,
	}

	auth := ctx.Engine.Group("/api/v1/auth")
	{
		auth.POST("/login", h.Login)
	}

	users := ctx.Engine.Group("/api/v1/users")
	users.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		users.POST("",
			middleware.RequirePermission(ctx.Queries, "create_user"),
			h.InsertUserInfo,
		)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req, h.keySecret)
	if err != nil {
		if err.Error() == "auth.login_failed" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Login failed: invalid username or password",
			})
			return
		}

		h.logger.Error("Login error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Login successful",
		"data":    token,
	})
}

func (h *AuthHandler) InsertUserInfo(c *gin.Context) {
	var req request.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.userService.InsertUserInfo(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to insert user", zap.Error(err))

		if err.Error() == "user.username_already_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"status":  http.StatusConflict,
				"message": "Username already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "User created successfully",
		"data": response.UserInfoResponse{
			UserID:    user.UserID,
			Username:  user.Username,
			Email:     user.Email,
			Status:    user.Status,
			Phone:     user.Phone,
			Address:   user.Address,
			CCCD:      user.Cccd,
			RoleID:    user.RoleID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

