package middleware

import (
	"IFMS-BE-API/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type confirmPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

func ConfirmPasswordMiddleware(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req confirmPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Password confirmation required",
			})
			return
		}

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "User not found",
			})
			return
		}

		if !utils.IsPasswordMatch(req.Password, user.PasswordHash) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Incorrect password",
			})
			return
		}

		c.Next()
	}
}
