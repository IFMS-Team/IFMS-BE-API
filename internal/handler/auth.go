package handler

import (
	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/middleware"
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		users.GET("/me", h.GetMyProfile)
		users.GET("",
			middleware.RequireRole(ctx.Queries, "admin", "sub_admin"),
			h.ListUsers,
		)
		users.GET("/search",
			middleware.RequireRole(ctx.Queries, "admin", "sub_admin"),
			h.GetUserByUsername,
		)
		users.GET("/:id",
			middleware.RequireRole(ctx.Queries, "admin", "sub_admin"),
			h.GetUserByID,
		)
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

func (h *AuthHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	users, total, err := h.userService.ListUsersWithRole(c.Request.Context(), int32(size), int32(offset))
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    users,
		"pagination": gin.H{
			"page":  page,
			"size":  size,
			"total": total,
		},
	})
}

func (h *AuthHandler) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "User not authenticated",
		})
		return
	}

	uid, ok := userID.(pgtype.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		if err.Error() == "user.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
			return
		}
		h.logger.Error("Failed to get profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
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

func (h *AuthHandler) GetUserByUsername(c *gin.Context) {
	username := strings.ToLower(strings.TrimSpace(c.Query("username")))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Username is required",
		})
		return
	}

	user, err := h.userService.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		if err.Error() == "user.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
			return
		}
		h.logger.Error("Failed to get user by username", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
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

func (h *AuthHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		if err.Error() == "user.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
			return
		}
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
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
