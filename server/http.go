package server

import (
	_ "gin-gorm-api/docs" // Required by swaggo/gin-swagger.
	"gin-gorm-api/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewEngine() *gin.Engine {
	r := gin.Default()
	r.GET("/hello", handler.HelloHandler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
