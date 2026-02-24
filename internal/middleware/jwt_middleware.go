package middleware

import (
	"net/http"
	"strings"

	"github.com/cachatto/master-slave-server/internal/service"
	"github.com/gin-gonic/gin"
)

// JWTAuth returns a Gin middleware that validates JWT access tokens.
// It extracts the token from the Authorization header (Bearer <token>),
// validates it, and sets "userID" and "email" in the Gin context.
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must be in the format: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]

		claims, err := authService.ParseAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Set user info in the context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
