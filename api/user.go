package api

import (
	"errors"
	"gin-gorm-api/middleware"
	"gin-gorm-api/model"
	"gin-gorm-api/schema"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHanlder exposes endpoints to interact with the User model.
type UserHanlder struct {
	db     *gorm.DB
	authMW gin.HandlerFunc
}

// NewUserHandler returns a new UserHanlder.
func NewUserHandler(db *gorm.DB, authMW gin.HandlerFunc) UserHanlder {
	return UserHanlder{db: db, authMW: authMW}
}

// CreateUser godoc
// @Summary      Create user
// @Schemes
// @Description  Create new user
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        form     body      schema.NewUserForm true "User form"
// @Success      201      {object}  schema.UserOut
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      409      {object}  schema.Errors "Duplicate user"
// @Failure      default  {string}  string        "Unexpected error"
// @Router       /user/   [post]
// .
func (h UserHanlder) create(c *gin.Context) {
	formData, _ := c.Get("form")
	form, _ := formData.(schema.NewUserForm)
	user := model.User{Username: form.Username, Email: form.Email}
	if err := user.SetPassword(form.Password); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r := h.db.WithContext(c.Request.Context()).Create(&user)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, schema.SimpleError(r.Error))
			return
		}
		_ = c.AbortWithError(http.StatusFailedDependency, r.Error)
		return
	}
	c.JSON(
		http.StatusCreated,
		schema.UserOut{ID: user.ID, Username: user.Username, Email: user.Email},
	)
}

// GetUsers godoc
// @Summary      Get all users
// @Schemes
// @Description  Get all users
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200      {object}  []schema.UserOut
// @Failure      403
// @Failure      default  {string}  string "Unexpected error"
// @Router       /user/   [get]
// .
func (h UserHanlder) getAll(c *gin.Context) {
	var users []schema.UserOut
	if r := h.db.WithContext(c.Request.Context()).Model(&model.User{}).Find(
		&users,
	); r.Error != nil {
		_ = c.AbortWithError(http.StatusFailedDependency, r.Error)
	}
	c.JSON(http.StatusOK, users)
}

// GetUserById godoc
// @Summary      Get user
// @Schemes
// @Description  Get user by ID
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id  path      int true "User id"
// @Success      200      {object}  schema.UserOut
// @Failure      400      {object}  schema.Errors "Bad request"
// @Failure      403
// @Failure      404
// @Failure      default  {string}  string "Unexpected error"
// @Router       /user/{user_id}   [get]
// .
func (h UserHanlder) getByID(c *gin.Context) {
	var user schema.UserOut
	userID, err := getParamID("userid", c)
	if err != nil {
		c.JSON(http.StatusBadRequest, schema.Errors{"user_id": err.Error()})
		return
	}

	r := h.db.WithContext(c.Request.Context()).Model(&model.User{}).First(
		&user,
		userID,
	)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusFailedDependency, r.Error)
		return
	}
	c.JSON(http.StatusOK, user)
}

// AddRoutes add a group of routes to r under the path "/user".
func (h UserHanlder) AddRoutes(r *gin.Engine) {
	g := r.Group("/user")
	g.POST("/", middleware.FormValidation[schema.NewUserForm](), h.create)
	g.GET("/", h.authMW, h.getAll)
	g.GET("/:userid", h.authMW, h.getByID)
}
