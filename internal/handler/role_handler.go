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
		roles.GET("",
			middleware.RequirePermission(ctx.Queries, "view_roles"),
			h.ListRoles,
		)
		roles.GET("/:id",
			middleware.RequirePermission(ctx.Queries, "view_roles"),
			h.GetRoleWithPermissions,
		)
		roles.POST("",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.CreateRole,
		)
		roles.PUT("/:id",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.UpdateRole,
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
		perms.GET("",
			middleware.RequirePermission(ctx.Queries, "view_roles"),
			h.ListPermissions,
		)
		perms.POST("",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.CreatePermission,
		)
		perms.PUT("/:id",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.UpdatePermission,
		)
		perms.DELETE("/:id",
			middleware.RequirePermission(ctx.Queries, "manage_roles"),
			h.DeletePermission,
		)
	}
}

// ListRoles godoc
// @Summary      List roles
// @Description  Get all roles (requires view_roles permission)
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.APIResponse{data=[]response.RoleResponse}
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles [get]
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

// GetRoleWithPermissions godoc
// @Summary      Get role with permissions
// @Description  Get role details including assigned permissions
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Role UUID"
// @Success      200 {object} response.APIResponse{data=response.RoleWithPermissionsResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /api/v1/roles/{id} [get]
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

// CreateRole godoc
// @Summary      Create role
// @Description  Create a new role (requires manage_roles permission)
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body request.CreateRoleRequest true "Role info"
// @Success      201 {object} response.APIResponse{data=response.RoleResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse "Role already exists"
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles [post]
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

// UpdateRole godoc
// @Summary      Update role
// @Description  Update an existing role by ID (requires manage_roles permission)
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Role UUID"
// @Param        body body request.UpdateRoleRequest true "Updated role info"
// @Success      200 {object} response.APIResponse{data=response.RoleResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid role ID",
		})
		return
	}

	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	role, err := h.roleService.UpdateRole(c.Request.Context(), roleID, req)
	if err != nil {
		if err.Error() == "role.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Role not found",
			})
			return
		}
		h.logger.Error("Failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Role updated",
		"data":    response.ToRoleResponse(role),
	})
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Delete a role by ID (requires manage_roles permission)
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Role UUID"
// @Success      200 {object} response.MessageResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles/{id} [delete]
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

// AssignPermission godoc
// @Summary      Assign permission to role
// @Description  Assign a permission to a specific role (requires manage_roles permission)
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Role UUID"
// @Param        body body request.AssignPermissionRequest true "Permission to assign"
// @Success      200 {object} response.MessageResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles/{id}/permissions [post]
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

// RemovePermission godoc
// @Summary      Remove permission from role
// @Description  Remove a specific permission from a role (requires manage_roles permission)
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Role UUID"
// @Param        permId path string true "Permission UUID"
// @Success      200 {object} response.MessageResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/roles/{id}/permissions/{permId} [delete]
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

// ListPermissions godoc
// @Summary      List permissions
// @Description  Get all permissions (requires view_roles permission)
// @Tags         Permissions
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.APIResponse{data=[]response.PermissionResponse}
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/permissions [get]
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

// CreatePermission godoc
// @Summary      Create permission
// @Description  Create a new permission (requires manage_roles permission)
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body request.CreatePermissionRequest true "Permission info"
// @Success      201 {object} response.APIResponse{data=response.PermissionResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/permissions [post]
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

// UpdatePermission godoc
// @Summary      Update permission
// @Description  Update an existing permission by ID (requires manage_roles permission)
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Permission UUID"
// @Param        body body request.UpdatePermissionRequest true "Updated permission info"
// @Success      200 {object} response.APIResponse{data=response.PermissionResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/permissions/{id} [put]
func (h *RoleHandler) UpdatePermission(c *gin.Context) {
	permID, err := response.StringToUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid permission ID",
		})
		return
	}

	var req request.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	perm, err := h.roleService.UpdatePermission(c.Request.Context(), permID, req)
	if err != nil {
		if err.Error() == "permission.not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Permission not found",
			})
			return
		}
		h.logger.Error("Failed to update permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Permission updated",
		"data":    response.ToPermissionResponse(perm),
	})
}

// DeletePermission godoc
// @Summary      Delete permission
// @Description  Delete a permission by ID (requires manage_roles permission)
// @Tags         Permissions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Permission UUID"
// @Success      200 {object} response.MessageResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/permissions/{id} [delete]
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
