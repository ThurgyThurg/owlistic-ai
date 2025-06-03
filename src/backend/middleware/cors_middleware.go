package middleware

import (
	"strings"
	gin "github.com/gin-gonic/gin"
)

// CORSMiddleware adds the required headers to allow cross-origin requests
// with proper handling for credentials and specific origins
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	// Parse allowed origins from comma-separated string
	origins := strings.Split(allowedOrigins, ",")
	originMap := make(map[string]bool)
	hasWildcard := false
	
	for _, origin := range origins {
		trimmedOrigin := strings.TrimSpace(origin)
		if trimmedOrigin == "*" {
			hasWildcard = true
		} else if trimmedOrigin != "" {
			originMap[trimmedOrigin] = true
		}
	}
	
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Determine if we should allow this origin
		allowOrigin := false
		if originMap[origin] {
			// Specific origin is allowed
			allowOrigin = true
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if hasWildcard && origin == "" {
			// Allow wildcard for non-browser requests (e.g., curl, Postman)
			c.Header("Access-Control-Allow-Origin", "*")
		} else if hasWildcard {
			// For browser requests with wildcard, we can't use credentials
			c.Header("Access-Control-Allow-Origin", "*")
			// Note: Credentials won't work with wildcard
		} else if len(originMap) == 0 {
			// No origins configured, allow all for development
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		// Always set these headers
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Requested-With, Origin, Cache-Control, X-File-Name, Cf-Access-Jwt-Assertion, Cf-Access-Authenticated-User-Email")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "43200") // 12 hours
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if allowOrigin || hasWildcard || len(originMap) == 0 {
				c.AbortWithStatus(200)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}
		
		c.Next()
	})
}

// CloudflareAccessMiddleware validates Cloudflare Access headers (optional)
func CloudflareAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cloudflare Access headers
		cfEmail := c.GetHeader("Cf-Access-Authenticated-User-Email")
		cfJWT := c.GetHeader("Cf-Access-Jwt-Assertion")
		
		if cfEmail != "" {
			// User authenticated via Cloudflare Access
			c.Set("cf-email", cfEmail)
			c.Set("cf-jwt", cfJWT)
			
			// You could add additional validation here:
			// - Verify the JWT signature using Cloudflare's public keys
			// - Check if the email is in your allowed list
			// - Log access for audit purposes
		}
		
		c.Next()
	}
}