package handler

import (
	"net/http"
	"regexp"

	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/middleware"
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"
	"IFMS-BE-API/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type FloorHandler struct {
	logger  *zap.Logger
	service *services.FloorService
}

var floorNameRegex = regexp.MustCompile(`^[A-Za-z0-9 _\-]{1,100}$`)

func NewFloorHandler(ctx *app.AppContext) {
	floorRepo := repository.NewFloorRepository(ctx.Pool, ctx.Queries)
	buildingRepo := repository.NewBuildingRepository(ctx.Pool)
	floorService := services.NewFloorService(floorRepo, buildingRepo, ctx.Logger)

	h := &FloorHandler{
		logger:  ctx.Logger,
		service: floorService,
	}

	floors := ctx.Engine.Group("/api/v1/floors")
	// Require Authentication first, then roles
	floors.Use(middleware.AuthMiddleware(ctx.KeySecret))
	floors.Use(middleware.RequireRole(ctx.Queries, "Admin", "Sub-admin"))
	{
		floors.POST("", h.CreateFloor)
		floors.PUT("/:id", h.UpdateFloor)
	}
}

func (h *FloorHandler) CreateFloor(c *gin.Context) {
	var req request.CreateFloorRequest

	// 1. Payload validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Missing required fields or invalid format",
		})
		return
	}

	if req.BuildingID == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Missing required fields buildingId and name",
		})
		return
	}

	if !floorNameRegex.MatchString(req.Name) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  http.StatusUnprocessableEntity,
			"message": "Regex mismatch for name",
		})
		return
	}

	userIDAny, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"message": "User not authenticated",
		})
		return
	}

	userID, ok := userIDAny.(pgtype.UUID)
	if !ok || !userID.Valid {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"message": "Invalid user ID",
		})
		return
	}

	// Create Floor
	resp, err := h.service.Create(c.Request.Context(), req, userID)
	if err != nil {
		if err == services.ErrBuildingNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Building not found",
			})
			return
		}
		if err == services.ErrBuildingMaxCapacity {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  http.StatusUnprocessableEntity,
				"message": err.Error(),
			})
			return
		}
		
		h.logger.Error("Failed to create floor internal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create floor",
		})
		return
	}

	// Success
	c.JSON(http.StatusCreated, resp)
}

func (h *FloorHandler) UpdateFloor(c *gin.Context) {
	floorID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid floor ID format",
		})
		return
	}

	var req request.UpdateFloorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	userIDAny, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "User not authenticated",
		})
		return
	}
	userID, ok := userIDAny.(pgtype.UUID)
	if !ok || !userID.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "Invalid user ID",
		})
		return
	}

	resp, err := h.service.Update(c.Request.Context(), floorID, req, userID)
	if err != nil {
		if err.Error() == "floor not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Floor not found",
			})
			return
		}
		h.logger.Error("Failed to update floor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update floor",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Floor updated",
		"data":    resp,
	})
}
