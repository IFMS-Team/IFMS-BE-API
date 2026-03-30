package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

func RequireRole(queries *db.Queries, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			return
		}

		uid, ok := userID.(pgtype.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Invalid user ID",
			})
			return
		}

		user, err := queries.GetUserByID(c.Request.Context(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "User not found",
			})
			return
		}

		role, err := queries.GetRoleByID(c.Request.Context(), user.RoleID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Role not found",
			})
			return
		}

		for _, allowed := range allowedRoles {
			if strings.EqualFold(role.RoleName, allowed) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"message": "Your role does not have access to this resource",
		})
	}
}

func RequirePermission(queries *db.Queries, permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			return
		}

		uid, ok := userID.(pgtype.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Invalid user ID",
			})
			return
		}

		user, err := queries.GetUserByID(c.Request.Context(), uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "User not found",
			})
			return
		}

		hasPermission, err := queries.CheckRoleHasPermission(c.Request.Context(), db.CheckRoleHasPermissionParams{
			RoleID:         user.RoleID,
			PermissionName: permissionName,
		})
		if err != nil || !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "You don't have permission to perform this action",
			})
			return
		}

		c.Next()
	}
}
