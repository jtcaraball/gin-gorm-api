package middleware

import (
	"gin-gorm-api/provider"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewSessionMiddleware returns a middleware that verifies if a session exists
// and if so adds the corresponding user to c under the key manager.UserKey.
// Authentication is handled by the given manager which is espected inmutable.
func NewSessionMiddleware(manager provider.UserAuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := manager.RetrieveSession(c)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Set(manager.UserKey, user)
		c.Next()
	}
}
