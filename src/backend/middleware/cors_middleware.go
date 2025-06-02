package middleware

import (
	gin "github.com/gin-gonic/gin"
)

// CORSMiddleware adds the required headers to allow cross-origin requests
func CORSMiddleware(AppOrigins string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Allow all origins for now - can be restricted later
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Requested-With, Origin, Cache-Control, X-File-Name")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "43200") // 12 hours
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}
