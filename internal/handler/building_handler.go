package handler

import (
	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/middleware"
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"
	"IFMS-BE-API/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

type BuildingHandler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
	service *services.BuildingService
}

func NewBuildingHandler(ctx *app.AppContext) {
	buildingRepo := repository.NewBuildingRepository(ctx.Pool)
	buildingService := services.NewBuildingService(buildingRepo, ctx.Logger)

	h := &BuildingHandler{
		pool:    ctx.Pool,
		queries: ctx.Queries,
		logger:  ctx.Logger,
		service: buildingService,
	}

	buildings := ctx.Engine.Group("/api/v1/buildings")
	buildings.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		buildings.GET("",
			h.ListBuildings,
		)
		buildings.GET("/:id",
			h.GetBuilding,
		)
		buildings.GET("/:id/floors",
			h.ListFloorsByBuilding,
		)
		buildings.POST("",
			middleware.RequirePermission(ctx.Queries, "create_building"),
			h.CreateBuilding,
		)
		buildings.PUT("/:id",
			middleware.RequirePermission(ctx.Queries, "create_building"),
			h.UpdateBuilding,
		)
	}
}

func (h *BuildingHandler) ListBuildings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size
	data, total, err := h.service.List(c.Request.Context(), int32(size), int32(offset))
	if err != nil {
		h.logger.Error("Failed to list buildings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get buildings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    data,
		"pagination": gin.H{
			"page":  page,
			"size":  size,
			"total": total,
		},
	})
}

func (h *BuildingHandler) ListFloorsByBuilding(c *gin.Context) {
	buildingID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid building ID",
		})
		return
	}

	floors, err := h.queries.ListFloorsByBuildingID(c.Request.Context(), buildingID)
	if err != nil {
		h.logger.Error("Failed to list floors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get floors",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    response.ToFloorListResponse(floors),
	})
}

func (h *BuildingHandler) CreateBuilding(c *gin.Context) {
	var req request.CreateBuildingRequest
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

	resp, err := h.service.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create building",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Building created",
		"data":    resp,
	})
}

func (h *BuildingHandler) UpdateBuilding(c *gin.Context) {
	buildingID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid building ID format",
		})
		return
	}

	var req request.UpdateBuildingRequest
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

	resp, err := h.service.Update(c.Request.Context(), buildingID, req, userID)
	if err != nil {
		if err.Error() == "building not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Building not found",
			})
			return
		}
		h.logger.Error("Failed to update building", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update building",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Building updated",
		"data":    resp,
	})
}

func (h *BuildingHandler) GetBuilding(c *gin.Context) {
	idStr := c.Param("id")
	var buildingID pgtype.UUID
	if err := buildingID.Scan(idStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid building ID format",
		})
		return
	}

	building, err := h.service.GetByID(c.Request.Context(), buildingID)
	if err != nil {
		if err.Error() == "building not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Building not found",
			})
			return
		}
		h.logger.Error("Failed to get building", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get building",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    building,
	})
}
