package server

import (
	"fmt"
	_ "gin-gorm-api/docs" // Required by swaggo/gin-swagger.
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// A routeAdder is able to add routes to a gin.Engine.
type routeAdder interface {
	AddRoutes(g *gin.Engine)
}

// allowedHosts returns a middleware that confirms that a given request's host
// name is in ah.
func allowedHosts(ah []string) gin.HandlerFunc {
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

// headers returns a middleware that attaches default headers to a request.
func headers() gin.HandlerFunc {
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

// initEngine return a base gin.Engine as specified by config.
func initEngine(config Config) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(headers())
	r.Use(allowedHosts(config.Engine.AllowedHost))
	err := r.SetTrustedProxies(config.Engine.TrustedProxies)
	if err != nil {
		return nil, err
	}
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return r, nil
}

// NewEngine returns a gin.Engine with the routes added by handlers.
func NewEngine(config Config, handlers ...routeAdder) (*gin.Engine, error) {
	r, err := initEngine(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}
	for _, handler := range handlers {
		handler.AddRoutes(r)
	}
	// Add swagger 2.0 spec.
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r, nil
}
