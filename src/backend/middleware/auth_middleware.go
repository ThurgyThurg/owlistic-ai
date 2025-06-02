package middleware

import (
	"net/http"

	"owlistic-notes/owlistic/services"
	"owlistic-notes/owlistic/utils/token"

	"github.com/gin-gonic/gin"
)

// ExtractAndValidateToken uses the token utility instead
func ExtractAndValidateToken(c *gin.Context, authService services.AuthServiceInterface) (*token.JWTClaims, error) {
	// Extract token from query or header
	tokenString, err := token.ExtractToken(c)
	if err != nil {
		return nil, err
	}

	// Validate the token using the auth service (which now uses the token utility)
	return authService.ValidateToken(tokenString)
}

func AuthMiddleware(authService services.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Extract and validate token
		claims, err := ExtractAndValidateToken(c, authService)
		if err != nil {
			// Add debug logging
			c.Header("X-Debug-Auth-Error", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Store user info in the context for later use
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
