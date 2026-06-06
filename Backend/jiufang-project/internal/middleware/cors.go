// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// CORSConfig contains the configuration for CORS middleware.
type CORSConfig struct {
	// AllowOrigins is the list of allowed origins
	AllowOrigins []string

	// AllowMethods is the list of allowed HTTP methods
	AllowMethods []string

	// AllowHeaders is the list of allowed HTTP headers
	AllowHeaders []string

	// ExposeHeaders is the list of headers that can be exposed to the client
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int
}

// DefaultCORSConfig returns the default CORS configuration for local development.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173",  // Vite dev server
			"http://localhost:3000",  // Alternative frontend port
			"http://127.0.0.1:5173", // Vite dev server (IP)
			"http://127.0.0.1:3000", // Alternative frontend port (IP)
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"PATCH",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
	}
}

// CORS returns a CORS middleware handler.
func CORS(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowOrigin := range config.AllowOrigins {
			if origin == allowOrigin {
				allowed = true
				break
			}
		}

		// If origin is not in the allowed list, return without CORS headers
		if !allowed && origin != "" {
			c.Next()
			return
		}

		// Set CORS headers
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowOrigins) > 0 {
			// If no origin header, use the first allowed origin
			c.Header("Access-Control-Allow-Origin", config.AllowOrigins[0])
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", boolToString(config.AllowCredentials))
		c.Header("Access-Control-Max-Age", intToString(config.MaxAge))

		// Handle OPTIONS request (preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += sep
		}
		result += str
	}
	return result
}

// boolToString converts a bool to a string.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// intToString converts an int to a string.
func intToString(i int) string {
	return strconv.Itoa(i)
}
