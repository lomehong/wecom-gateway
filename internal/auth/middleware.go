package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKey is the key used to store auth context in Gin context
	ContextKey = "auth_context"
)

// GinMiddleware returns a Gin middleware for API key authentication
func GinMiddleware(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from Authorization header
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

		apiKey := strings.TrimSpace(parts[1])
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "empty api key",
			})
			c.Abort()
			return
		}

		// Authenticate
		authCtx, err := authenticator.Authenticate(c.Request.Context(), apiKey)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidAPIKey):
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    40101,
					"message": "invalid or missing API key",
				})
			case errors.Is(err, ErrAPIKeyDisabled):
				c.JSON(http.StatusForbidden, gin.H{
					"code":    40301,
					"message": "API key is disabled",
				})
			case errors.Is(err, ErrAPIKeyExpired):
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    40101,
					"message": "API key has expired",
				})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    40101,
					"message": "authentication failed",
				})
			}
			c.Abort()
			return
		}

		// Store auth context in Gin context
		c.Set(ContextKey, authCtx)
		c.Next()
	}
}

// RequirePermission returns a middleware that checks for required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := c.Get(ContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		ac, ok := authCtx.(*AuthContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "invalid auth context",
			})
			c.Abort()
			return
		}

		if !ac.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "permission denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission returns a middleware that checks for any of the required permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := c.Get(ContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		ac, ok := authCtx.(*AuthContext)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "invalid auth context",
			})
			c.Abort()
			return
		}

		if !ac.HasAnyPermission(permissions) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "permission denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin returns a middleware that requires admin privileges
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := c.Get(ContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		ac, ok := authCtx.(*AuthContext)
		if !ok || !ac.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAuthContext retrieves the auth context from Gin context
func GetAuthContext(c *gin.Context) (*AuthContext, bool) {
	authCtx, exists := c.Get(ContextKey)
	if !exists {
		return nil, false
	}

	ac, ok := authCtx.(*AuthContext)
	return ac, ok
}
