package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getParamID(key string, c *gin.Context) (int, error) {
	val := c.Param(key)
	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid id '%s'", val)
	}
	return id, nil
}
