package api

import (
	"fmt"
	"gin-gorm-api/config"
	_ "gin-gorm-api/docs" // Required by swaggo/gin-swagger.
	"gin-gorm-api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// initEngine return a base gin.Engine as specified by config.
func initEngine(conf config.Config) (*gin.Engine, error) {
	if conf.Testing {
		gin.SetMode(gin.TestMode)
	} else if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	var r *gin.Engine
	if !conf.Testing { // We don't want logging dirtying test outputs.
		r = gin.Default()
	} else {
		r = gin.New()
		r.Use(gin.Recovery())
	}

	r.Use(middleware.PolicyHeaders())
	r.Use(middleware.AllowedHosts(conf.Engine.AllowedHost))
	err := r.SetTrustedProxies(conf.Engine.TrustedProxies)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// NewEngine returns a gin.Engine with the routes added by handlers.
func NewEngine(
	conf config.Config,
) (*gin.Engine, error) {
	r, err := initEngine(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}
	// Add swagger 2.0 spec.
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r, nil
}
