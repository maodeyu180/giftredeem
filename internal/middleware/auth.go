package middleware

import (
	"giftredeem/internal/auth"
	"giftredeem/internal/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies JWT tokens and adds user information to the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")

		// Check if header exists and has Bearer token
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "Authorization header missing or invalid"))
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusOK, response.Error(response.CodeAuthInvalidToken, "Invalid token: "+err.Error()))
			c.Abort()
			return
		}

		// Get user from token
		user, err := auth.GetUserFromToken(claims)
		if err != nil {
			code := response.CodeUnauthorized
			if err == auth.ErrUserBanned {
				code = response.CodeAuthUserBanned
			}
			c.JSON(http.StatusOK, response.Error(code, "User authentication failed: "+err.Error()))
			c.Abort()
			return
		}

		// Store user and claims in context for later use
		c.Set("user", user)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware attempts to authenticate the user but allows requests to proceed if authentication fails
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")

		// Skip if no auth header or not a bearer token
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Try to validate token
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			// Continue without authentication
			c.Next()
			return
		}

		// Try to get user from token
		user, err := auth.GetUserFromToken(claims)
		if err == nil {
			// Store user and claims in context
			c.Set("user", user)
			c.Set("claims", claims)
		}

		c.Next()
	}
}
