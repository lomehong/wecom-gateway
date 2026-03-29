package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/httputil"
)

const (
	// JWTContextKey is the key used to store JWT claims in Gin context
	JWTContextKey = "jwt_claims"
)

// JWTMiddleware creates a JWT authentication middleware
func (am *AuthManager) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "empty token",
			})
			c.Abort()
			return
		}

		// Verify token
		claims, err := am.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set(JWTContextKey, claims)
		c.Next()
	}
}

// RequireAuth is a simple middleware that requires JWT authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(JWTContextKey)
		if !exists {
			httputil.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		_, ok := claims.(*JWTClaims)
		if !ok {
			httputil.Unauthorized(c, "Invalid authentication context")
			c.Abort()
			return
		}

		c.Next()
	}
}
