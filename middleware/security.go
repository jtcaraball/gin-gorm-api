package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// AllowedHosts returns a middleware that confirms that a given request's host
// name is in ah.
func AllowedHosts(ah []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// There should not be many allowedHosts so slices.Contains should be
		// faster than a hash table.
		if slices.Contains(ah, c.Request.Host) {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid host header"},
			)
			return
		}
		c.Next()
	}
}

// PolicyHeaders returns a middleware that attaches default headers to a
// request.
func PolicyHeaders() gin.HandlerFunc {
	// Following recommendations from:
	// https://github.com/gin-gonic/examples/tree/master/secure-web-app
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header(
			"Content-Security-Policy",
			"default-src 'self'; connect-src *; font-src *; "+
				"script-src-elem * 'unsafe-inline'; img-src * data:;"+
				"style-src * 'unsafe-inline';",
		)
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header(
			"Strict-Transport-Security",
			"max-age=31536000; includeSubDomains; preload",
		)
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header(
			"Permissions-Policy",
			"geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),"+
				"magnetometer=(),gyroscope=(),fullscreen=(self),payment=()",
		)
		c.Next()
	}
}
