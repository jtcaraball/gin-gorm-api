package server

import (
	"fmt"
	_ "gin-gorm-api/docs" // Required by swaggo/gin-swagger.

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// A routeAdder is able to add routes to a gin.Engine.
type routeAdder interface {
	AddToRouter(g *gin.Engine)
}

// initEngine return a base gin.Engine as specified by config.
func initEngine(config Config) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	if len(config.Engine.TrustedProxies) == 0 {
		return r, nil
	}
	err := r.SetTrustedProxies(config.Engine.TrustedProxies)
	return r, err
}

// NewEngine returns a gin.Engine with the routers added by handlers.
func NewEngine(config Config, handlers ...routeAdder) (*gin.Engine, error) {
	r, err := initEngine(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}
	for _, handler := range handlers {
		handler.AddToRouter(r)
	}
	// Add swagger 2.0 spec.
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r, nil
}
