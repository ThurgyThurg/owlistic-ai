package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
)

// CORSMiddleware adds the required headers to allow cross-origin requests
func CORSMiddleware(AppOrigins string) gin.HandlerFunc {

	// Set up CORS configuration
	corsConfig := cors.DefaultConfig()
	
	// Split and clean up origins
	origins := strings.Split(AppOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	
	corsConfig.AllowOrigins = origins
	corsConfig.AllowWildcard = true
	corsConfig.AllowWebSockets = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
	}
	corsConfig.AllowHeaders = []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"X-Requested-With",
		"Origin",
		"Cache-Control",
		"X-File-Name",
	}
	corsConfig.ExposeHeaders = []string{
		"Content-Length",
		"Content-Type",
	}
	corsConfig.MaxAge = 12 * 60 * 60 // 12 hours

	return cors.New(corsConfig)
}
