package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Greeter struct{}

// @Summary Welcome!
// @Description Say hello.
// @Produce  json
// @Success 200 {string} string	"ok"
// @Router /hello [get]
// .
func (g Greeter) Hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello, World!")
}

// Add handler methods to engine r's routes.
func (g Greeter) AddToRouter(r *gin.Engine) {
	r.GET("/hello", g.Hello)
}
