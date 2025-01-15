package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// A Form is any structure that can be bind from JSON and is able to be
// validated.
type form interface {
	Validate() error
}

// An authManager is able to execute basic authentication tasks.
type authManager[L, Q, R, C form, U any] interface {
	Authenticate(L) (U, bool)
	RegisterSession(U, *gin.Context)
	RemoveSession(U, *gin.Context)
	RetrieveSession(*gin.Context) (U, bool)
	RequestPasswordReset(Q) error
	ResetPassword(R) error
	SetPassword(U, C) error
}

// Service for all your authentication needs. As long as those are an
// authentication middleware and basic endpoints.
type Service[L, Q, R, C form, U any] struct {
	key     string
	manager authManager[L, Q, R, C, U]
}

// New authentication service powered by manager. The key string is used to
// identify the session instance in a contex's values.
func New[L, Q, R, C form, U any](
	key string,
	manager authManager[L, Q, R, C, U],
) Service[L, Q, R, C, U] {
	return Service[L, Q, R, C, U]{key: key, manager: manager}
}

// AuthMiddleware verifies if a session exists and if so adds the corresponding
// instance to c under the key s.Key.
func (s Service[L, Q, R, C, U]) AuthMiddleware(c *gin.Context) {
	user, ok := s.manager.RetrieveSession(c)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Set("user", user)
	c.Next()
}

// Handler that implements authentication and password managing
// functionalities.
type Handler[L, Q, R, C form, U any] struct {
	key        string
	manager    authManager[L, Q, R, C, U]
	middleware gin.HandlerFunc
}

// Handler returns an authentication handler based on s' internal manager.
func (s Service[L, Q, R, C, U]) Handler() Handler[L, Q, R, C, U] {
	return Handler[L, Q, R, C, U]{
		key:        s.key,
		manager:    s.manager,
		middleware: s.AuthMiddleware,
	}
}

func (h Handler[L, Q, R, C, U]) login(c *gin.Context) {
	var form L
	if err := c.BindJSON(&form); err != nil {
		return
	}
	if err := form.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	session, ok := h.manager.Authenticate(form)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
	h.manager.RegisterSession(session, c)
}

func (h Handler[L, Q, R, C, U]) logout(c *gin.Context) {
	sessionData, _ := c.Get(h.key)
	session, ok := sessionData.(U)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
	h.manager.RemoveSession(session, c)
}

func (h Handler[L, Q, R, C, U]) me(c *gin.Context) {
	sessionData, _ := c.Get(h.key)
	session, ok := sessionData.(U)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
	c.JSON(http.StatusAccepted, session)
}

func (h Handler[L, Q, R, C, U]) requestPasswordReset(c *gin.Context) {
	var form Q
	if err := c.BindJSON(&form); err != nil {
		return
	}
	if err := form.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if err := h.manager.RequestPasswordReset(form); err != nil {
		c.Status(http.StatusFailedDependency)
		return
	}
	c.Status(http.StatusOK)
}

func (h Handler[L, Q, R, C, U]) resetPassword(c *gin.Context) {
	var form R
	if err := c.BindJSON(&form); err != nil {
		return
	}
	if err := form.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if err := h.manager.ResetPassword(form); err != nil {
		c.Status(http.StatusFailedDependency)
		return
	}
	c.Status(http.StatusOK)
}

func (h Handler[L, Q, R, C, U]) changePassword(c *gin.Context) {
	sessionData, _ := c.Get(h.key)
	session, ok := sessionData.(U)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	var form C
	if err := c.BindJSON(&form); err != nil {
		return
	}
	if err := form.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if err := h.manager.SetPassword(session, form); err != nil {
		c.Status(http.StatusFailedDependency)
		return
	}
	c.Status(http.StatusOK)
}

func (h Handler[L, Q, R, C, U]) AddToRouter(r *gin.Engine, path string) {
	g := r.Group(path)
	g.POST("/", h.login)
	g.DELETE("/", h.middleware, h.logout)
	g.GET("/me", h.middleware, h.me)
	g.POST("/request_password_reset", h.requestPasswordReset)
	g.POST("/reset_password", h.resetPassword)
	g.POST("/change_password", h.middleware, h.changePassword)
}
