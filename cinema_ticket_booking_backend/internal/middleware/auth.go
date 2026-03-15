package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
)

func RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated."})
			c.Abort()
			return
		}

		userRole, ok := role.(models.UserRole)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user role."})
			c.Abort()
			return
		}

		for _, r := range allowedRoles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Insufficient permissions."})
		c.Abort()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleAdmin)
}

func RequireUser() gin.HandlerFunc {
	return RequireRole(models.UserRoleUser, models.UserRoleAdmin)
}
