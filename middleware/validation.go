package middleware

import (
	"gin-gorm-api/schema"
	"net/http"

	"github.com/gin-gonic/gin"
)

// A form can be validated.
type form interface {
	Validate() (schema.Errors, error)
}

// FormValidation returns a middleware that validates and sets the request's
// body, of type V, to the key "form".
func FormValidation[V form]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var f V
		if err := c.BindJSON(&f); err != nil {
			return
		}
		valErrs, err := f.Validate()
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if valErrs != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, valErrs)
			return
		}
		c.Set("form", f)
		c.Next()
	}
}
