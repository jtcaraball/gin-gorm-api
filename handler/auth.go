package handler

import (
	"errors"
	"gin-gorm-api/middleware"
	"gin-gorm-api/model"
	"gin-gorm-api/provider"
	"gin-gorm-api/schema"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler exposes its manager's methods as endpoints.
type AuthHandler struct {
	manager provider.UserAuthManager
	authMW  gin.HandlerFunc
}

// NewAuthHandler returns a new AuthHandler.
func NewAuthHandler(
	manager provider.UserAuthManager,
	authMW gin.HandlerFunc,
) AuthHandler {
	return AuthHandler{manager, authMW}
}

// LoginSession godoc
// @Summary      Login
// @Schemes
// @Description  Start session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        form     body      schema.LoginForm true "Login form"
// @Success      200      {object}  model.User
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      403      {string}  string        "Forbidden"
// @Failure      default  {string}  string        "Unexpected error"
// @Router       /auth    [post]
// .
func (h AuthHandler) login(c *gin.Context) {
	formData, _ := c.Get("form")
	form, _ := formData.(schema.LoginForm)
	session, err := h.manager.Authenticate(form, c)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}
	if err = h.manager.RegisterSession(session, c); err != nil {
		_ = c.AbortWithError(http.StatusFailedDependency, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

// LogoutSession godoc
// @Summary      Logout
// @Schemes
// @Description  End session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      403      {string}  string  "forbidden"
// @Failure      default  {string}  string  "unexpected error"
// @Router       /auth    [delete]
// .
func (h AuthHandler) logout(c *gin.Context) {
	h.manager.RemoveSession(c)
}

// SessionMe godoc
// @Summary  Me
// @Schemes
// @Description  Current session information
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200      {object}  model.User
// @Failure      403      {string}  string  "forbidden"
// @Failure      default  {string}  string  "unexpected error"
// @Router       /auth/me [get]
// .
func (h AuthHandler) me(c *gin.Context) {
	sessionData, _ := c.Get(h.manager.UserKey)
	session, ok := sessionData.(model.User)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
	c.JSON(http.StatusAccepted, session)
}

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Schemes
// @Description  Request a password reset message
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        form     body      schema.PasswordResetRequestForm true "Password reset request form"
// @Success      200
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      404      {string}  string        "Email not found"
// @Failure      default  {string}  string        "Unexpected error"
// @Router       /auth/request_password_reset [post]
// .
func (h AuthHandler) requestPasswordReset(c *gin.Context) {
	formData, _ := c.Get("form")
	form, _ := formData.(schema.PasswordResetRequestForm)
	if err := h.manager.RequestPasswordReset(form, c); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusFailedDependency, err)
		return
	}
	c.Status(http.StatusOK)
}

// ResetPassword godoc
// @Summary      Password reset
// @Schemes
// @Description  Reset password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        form     body      schema.PasswordResetForm true "Password reset form"
// @Success      200
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      403      {string}  string        "Forbidden"
// @Failure      403      {object}  schema.Errors "Forbidden"
// @Failure      404      {string}  string        "Target not found"
// @Failure      default  {string}  string        "Unexpected error"
// @Router       /auth/reset_password [post]
// .
func (h AuthHandler) resetPassword(c *gin.Context) {
	formData, _ := c.Get("form")
	form, _ := formData.(schema.PasswordResetForm)
	if err := h.manager.ResetPassword(form, c); err != nil {
		handleTokenErrors(err, c)
		return
	}
	c.Status(http.StatusOK)
}

func handleTokenErrors(err error, c *gin.Context) {
	invalidToken := errors.Is(err, provider.ErrTokenExpired) ||
		errors.Is(err, provider.ErrInvalidToken)
	if invalidToken {
		c.JSON(http.StatusForbidden, schema.Errors{"error": err.Error()})
		return
	}
	c.Status(http.StatusForbidden)
}

// ChangePassword godoc
// @Summary      Change password
// @Schemes
// @Description  Change password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        form     body      schema.PasswordChangeForm true "Password change form"
// @Success      200
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      403      {string}  string        "Forbidden"
// @Failure      default  {string}  string        "Unexpected error"
// @Router       /auth/change_password [post]
// .
func (h AuthHandler) changePassword(c *gin.Context) {
	sessionData, _ := c.Get(h.manager.UserKey)
	session, ok := sessionData.(model.User)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
	formData, _ := c.Get("form")
	form, _ := formData.(schema.PasswordChangeForm)
	if err := h.manager.SetPassword(session, form, c); err != nil {
		_ = c.AbortWithError(http.StatusFailedDependency, err)
		return
	}
	c.Status(http.StatusOK)
}

// AddRoutes add a group of routes to r under the path "/auth".
func (h AuthHandler) AddRoutes(r *gin.Engine) {
	g := r.Group("/auth")
	g.POST("/", middleware.FormValidation[schema.LoginForm](), h.login)
	g.DELETE("/", h.authMW, h.logout)
	g.GET("/me", h.authMW, h.me)
	g.POST(
		"/request_password_reset",
		middleware.FormValidation[schema.PasswordResetRequestForm](),
		h.requestPasswordReset,
	)
	g.POST(
		"/reset_password",
		middleware.FormValidation[schema.PasswordResetForm](),
		h.resetPassword,
	)
	g.POST(
		"/change_password",
		h.authMW,
		middleware.FormValidation[schema.PasswordChangeForm](),
		h.changePassword,
	)
}
