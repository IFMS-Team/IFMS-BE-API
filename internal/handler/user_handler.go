package handler

import (
	"net/http"

	"IFMS-BE-API/internal/model"
	"IFMS-BE-API/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    user,
	})
}

func (h *UserHandler) List(c *gin.Context) {
	var pagination model.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid pagination",
			Error:   err.Error(),
		})
		return
	}

	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	users, err := h.service.List(c.Request.Context(), pagination.Page, pagination.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get users",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    users,
	})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "User created",
		Data:    user,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	user, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User updated",
		Data:    user,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User deleted",
	})
}
