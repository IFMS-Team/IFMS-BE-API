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
	userService := services.NewUserService(ctx.Pool, ctx.Queries, ctx.Logger, nil, ctx.Redis)

	h := &AuthHandler{
		queries:     ctx.Queries,
		keySecret:   ctx.KeySecret,
		logger:      ctx.Logger,
		userService: userService,
	}

	auth := ctx.Engine.Group("/api/v1/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/verify-otp", h.VerifyOTP)
		auth.POST("/reset-password", h.ResetPassword)
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
		users.PUT(
			"/:id",
			middleware.RequireRole(ctx.Queries, "admin", "sub_admin"),
			h.UpdateUserInfo,
		)
		users.PUT("/me/change-password", h.ChangePasswordByUser)
		users.PUT(
			"/:id/change-password",
			middleware.RequireRole(ctx.Queries, "admin", "sub_admin"),
			h.ChangePasswordByAdmin)
	}
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with username and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body request.LoginRequest true "Login credentials"
// @Success      200 {object} response.APIResponse{data=string} "JWT token"
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/auth/login [post]
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

// ForgotPassword godoc
// @Summary      Forgot password - request OTP
// @Description  Verify user identity (username, email, phone, cccd) and send OTP to email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body request.ForgotPasswordRequest true "Identity verification"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      429 {object} response.ErrorResponse "Max OTP attempts reached"
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	err := h.userService.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "user.identity_mismatch":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "The provided information does not match any account",
			})
		case "otp.max_attempts_reached":
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Too many attempts. Please try again later",
			})
		case "otp.send_failed":
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to send OTP. Please try again later",
			})
		case "service.redis_unavailable":
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  http.StatusServiceUnavailable,
				"message": "Service temporarily unavailable. Please try again later",
			})
		default:
			h.logger.Error("Forgot password error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "OTP has been sent to your registered email",
	})
}

// VerifyOTP godoc
// @Summary      Verify OTP
// @Description  Verify OTP code sent to email. Returns a reset token on success.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body request.VerifyOTPRequest true "OTP verification"
// @Success      200 {object} response.APIResponse{data=string} "Reset token"
// @Failure      400 {object} response.ErrorResponse
// @Failure      429 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req request.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	resetToken, err := h.userService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		errMsg := err.Error()
		switch {
		case errMsg == "otp.expired":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "OTP has expired. Please request a new one",
			})
		case errMsg == "otp.max_attempts_reached":
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Too many failed attempts. Please request a new OTP",
			})
		case strings.HasPrefix(errMsg, "otp.invalid:"):
			remaining := strings.TrimPrefix(errMsg, "otp.invalid:")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":    http.StatusBadRequest,
				"message":   "Invalid OTP code",
				"remaining": remaining,
			})
		default:
			h.logger.Error("Verify OTP error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      http.StatusOK,
		"message":     "OTP verified successfully",
		"reset_token": resetToken,
	})
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset password using the reset token from OTP verification
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body request.ResetPasswordRequest true "Reset password payload"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	err := h.userService.ResetPassword(c.Request.Context(), req)
	if err != nil {
		errMsg := err.Error()
		switch {
		case errMsg == "reset.token_expired":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Reset token has expired. Please start over",
			})
		case errMsg == "user.not_found":
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
		case errMsg == "auth.same_password":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "New password cannot be the same as the old password",
			})
		case strings.HasPrefix(errMsg, "password.weak:"):
			msg := strings.TrimPrefix(errMsg, "password.weak:")
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": msg,
			})
		default:
			h.logger.Error("Reset password error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to reset password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Password has been successfully reset",
	})
}

// UpdateUserInfo godoc
// @Summary      Update user
// @Description  Update user info by UUID (admin/sub_admin only). Username and CCCD cannot be modified.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string                    true "User UUID"
// @Param        body body request.UpdateUserRequest  true "Fields to update"
// @Success      200  {object} response.APIResponse{data=response.UserInfoResponse}
// @Failure      400  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      500  {object} response.ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *AuthHandler) UpdateUserInfo(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid user ID",
		})
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	if req.Username != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Username cannot be modified",
		})
		return
	}

	if req.CCCD != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "CCCD cannot be modified",
		})
		return
	}

	userID := pgtype.UUID{Bytes: uid, Valid: true}

	user, err := h.userService.UpdateUserInfo(c.Request.Context(), userID, req)
	if err != nil {
		if err.Error() == "user.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
			return
		}

		h.logger.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "User updated successfully",
		"data": response.UserInfoResponse{
			UserID:    user.UserID,
			Username:  user.Username,
			FullName:  user.FullName,
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

// ChangePasswordByAdmin godoc
// @Summary      Change password by admin
// @Description  Change the user's password by admin (admin/sub_admin only)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body request.ChangePasswordByAdminRequest true "Password change payload"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users/{id}/change-password [put]
func (h *AuthHandler) ChangePasswordByAdmin(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid user ID",
		})
		return
	}

	var req request.ChangePasswordByAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	userID := pgtype.UUID{Bytes: uid, Valid: true}

	err = h.userService.ChangePasswordByAdmin(c.Request.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "user.not_found":
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
		case "auth.same_password":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "New password must be different from current password",
			})
		default:
			h.logger.Error("Failed to change password by admin", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to change password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Password changed successfully",
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change the authenticated user's password. Requires old password verification.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body request.ChangePasswordByUserRequest true "Password change payload"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users/me/change-password [put]
func (h *AuthHandler) ChangePasswordByUser(c *gin.Context) {
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

	var req request.ChangePasswordByUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	err := h.userService.ChangePasswordByUser(c.Request.Context(), uid, req)
	if err != nil {
		switch err.Error() {
		case "user.not_found":
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "User not found",
			})
		case "auth.old_password_incorrect":
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Old password is incorrect",
			})
		case "auth.same_password":
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "New password must be different from old password",
			})
		default:
			h.logger.Error("Failed to change password", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to change password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Password changed successfully",
	})
}

// InsertUserInfo godoc
// @Summary      Create user
// @Description  Create a new user (requires create_user permission)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body request.CreateUserRequest true "User info"
// @Success      201 {object} response.APIResponse{data=response.UserInfoResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse "Username already exists"
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users [post]
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
			FullName:  user.FullName,
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

// ListUsers godoc
// @Summary      List users
// @Description  Get paginated list of users (admin/sub_admin only)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        size query int false "Page size" default(10)
// @Success      200 {object} response.PaginatedResponse{data=[]response.UserInfoResponse}
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users [get]
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

// GetMyProfile godoc
// @Summary      Get my profile
// @Description  Get the authenticated user's profile
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.APIResponse{data=response.UserInfoResponse}
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users/me [get]
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
			FullName:  user.FullName,
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

// GetUserByUsername godoc
// @Summary      Search user by username
// @Description  Find a user by username (admin/sub_admin only)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        username query string true "Username to search"
// @Success      200 {object} response.APIResponse{data=response.UserInfoResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users/search [get]
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
			FullName:  user.FullName,
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

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get user details by UUID (admin/sub_admin only)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Success      200 {object} response.APIResponse{data=response.UserInfoResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/users/{id} [get]
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
			FullName:  user.FullName,
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
