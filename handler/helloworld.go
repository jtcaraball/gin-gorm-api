package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Welcome!
// @Description Say hello.
// @Produce  json
// @Success 200 {string} string	"ok"
// @Router /hello [get]
// .
func HelloHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello, World!")
}
