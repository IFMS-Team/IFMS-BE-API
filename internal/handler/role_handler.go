package handler

import (
	"IFMS-BE-API/internal/app"
	"IFMS-BE-API/internal/middleware"
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"
	"IFMS-BE-API/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleHandler struct {
	roleService *services.RoleService
	logger      *zap.Logger
}

func NewRoleHandler(ctx *app.AppContext) {
	roleRepo := repository.NewRoleRepository(ctx.Pool)
	roleService := services.NewRoleService(roleRepo, ctx.Logger)

	h := &RoleHandler{
		roleService: roleService,
		logger:      ctx.Logger,
	}

	roles := ctx.Engine.Group("/api/v1/roles")
	roles.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		roles.GET("", h.ListRoles)
		roles.GET("/:id", h.GetRoleWithPermissions)
		roles.POST("",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.CreateRole,
		)
		roles.DELETE("/:id",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.DeleteRole,
		)
		roles.POST("/:id/permissions",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.AssignPermission,
		)
		roles.DELETE("/:id/permissions/:permId",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.RemovePermission,
		)
	}

	perms := ctx.Engine.Group("/api/v1/permissions")
	perms.Use(middleware.AuthMiddleware(ctx.KeySecret))
	{
		perms.GET("", h.ListPermissions)
		perms.POST("",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.CreatePermission,
		)
		perms.DELETE("/:id",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.DeletePermission,
		)
	}
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.roleService.ListRoles(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    response.ToRoleListResponse(roles),
	})
}

func (h *RoleHandler) GetRoleWithPermissions(c *gin.Context) {
	roleID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid role ID",
		})
		return
	}

	result, err := h.roleService.GetRoleWithPermissions(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Role not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    result,
	})
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req request.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "role.already_exists" {
			c.JSON(http.StatusConflict, gin.H{
				"status":  http.StatusConflict,
				"message": "Role already exists",
			})
			return
		}
		h.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create role",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Role created",
		"data":    response.ToRoleResponse(role),
	})
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid role ID",
		})
		return
	}

	if err := h.roleService.DeleteRole(c.Request.Context(), roleID); err != nil {
		if err.Error() == "role.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Role not found",
			})
			return
		}
		h.logger.Error("Failed to delete role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to delete role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Role deleted",
	})
}

func (h *RoleHandler) AssignPermission(c *gin.Context) {
	roleID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid role ID",
		})
		return
	}

	var req request.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	permID, err := response.StringToUUID(req.PermissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid permission ID",
		})
		return
	}

	if err := h.roleService.AssignPermission(c.Request.Context(), roleID, permID); err != nil {
		h.logger.Error("Failed to assign permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Permission assigned",
	})
}

func (h *RoleHandler) RemovePermission(c *gin.Context) {
	roleID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid role ID",
		})
		return
	}

	permID, err := response.StringToUUID(c.Param("permId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid permission ID",
		})
		return
	}

	if err := h.roleService.RemovePermission(c.Request.Context(), roleID, permID); err != nil {
		h.logger.Error("Failed to remove permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to remove permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Permission removed",
	})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.roleService.ListPermissions(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"data":    response.ToPermissionListResponse(perms),
	})
}

func (h *RoleHandler) CreatePermission(c *gin.Context) {
	var req request.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	perm, err := h.roleService.CreatePermission(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create permission",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Permission created",
		"data":    response.ToPermissionResponse(perm),
	})
}

func (h *RoleHandler) DeletePermission(c *gin.Context) {
	permID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid permission ID",
		})
		return
	}

	if err := h.roleService.DeletePermission(c.Request.Context(), permID); err != nil {
		h.logger.Error("Failed to delete permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to delete permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Permission deleted",
	})
}
