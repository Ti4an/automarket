package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"automarket/models"
	"automarket/services"
)

const userKey = "user"

// Authentication: ¿quién eres?
func Authenticate(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": services.ErrInvalidToken.Error()})
			return
		}
		user, err := auth.Authenticate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(userKey, user)
		c.Next()
	}
}

// Authorization: ¿tu rol te permite hacer esto?
func RequireRole(role models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c).Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": services.ErrForbidden.Error()})
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) models.User {
	value, _ := c.Get(userKey)
	user, _ := value.(models.User)
	return user
}