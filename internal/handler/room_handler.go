package handler

import (
	"net/http"

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

type RoomHandler struct {
	logger  *zap.Logger
	service *services.RoomService
}

func NewRoomHandler(ctx *app.AppContext) {
	floorRepo := repository.NewFloorRepository(ctx.Pool, ctx.Queries)
	roomRepo := repository.NewRoomRepository(ctx.Pool, ctx.Queries)
	roomService := services.NewRoomService(roomRepo, floorRepo, ctx.Logger)

	h := &RoomHandler{
		logger:  ctx.Logger,
		service: roomService,
	}

	rooms := ctx.Engine.Group("/api/v1/rooms")
	rooms.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		rooms.GET("/:id", h.GetRoom)
		rooms.POST("",
			middleware.RequireRole(ctx.Queries, "Admin", "Sub-admin"),
			h.CreateRoom,
		)
		rooms.PUT("/:id",
			middleware.RequireRole(ctx.Queries, "Admin", "Sub-admin"),
			h.UpdateRoom,
		)
	}

	floorRooms := ctx.Engine.Group("/api/v1/floors")
	floorRooms.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		floorRooms.GET("/:id/rooms", h.ListRoomsByFloor)
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req request.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Missing required fields or invalid format",
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
		switch err {
		case services.ErrFloorNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Floor not found",
			})
		case services.ErrFloorMaxCapacity:
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  http.StatusUnprocessableEntity,
				"message": err.Error(),
			})
		default:
			h.logger.Error("Failed to create room", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to create room",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Room created",
		"data":    resp,
	})
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	roomID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid room ID format",
		})
		return
	}

	var req request.UpdateRoomRequest
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

	resp, err := h.service.Update(c.Request.Context(), roomID, req, userID)
	if err != nil {
		if err.Error() == "room not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Room not found",
			})
			return
		}
		h.logger.Error("Failed to update room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Room updated",
		"data":    resp,
	})
}

func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid room ID format",
		})
		return
	}

	room, err := h.service.GetByID(c.Request.Context(), roomID)
	if err != nil {
		if err.Error() == "room not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Room not found",
			})
			return
		}
		h.logger.Error("Failed to get room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    room,
	})
}

func (h *RoomHandler) ListRoomsByFloor(c *gin.Context) {
	floorID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid floor ID format",
		})
		return
	}

	rooms, err := h.service.ListByFloorID(c.Request.Context(), floorID)
	if err != nil {
		h.logger.Error("Failed to list rooms", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get rooms",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    rooms,
	})
}
