package api_test

import (
	"encoding/json"
	"gin-gorm-api/api"
	"gin-gorm-api/schema"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newTestEngineWithUserHandler() (*gin.Engine, sqlmock.Sqlmock, error) {
	db, mock, err := newMockDB()
	if err != nil {
		return nil, nil, err
	}
	auth, sessionMW, err := newAuth(db)
	if err != nil {
		return nil, nil, err
	}
	r, err := api.NewEngine(
		testConf,
		api.NewAuthHandler(auth, sessionMW),
		api.NewUserHandler(db, sessionMW),
	)
	if err != nil {
		return nil, nil, err
	}
	return r, mock, nil
}

func TestUserHandler_create_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	mockInsertQuery(mock, &user, "^INSERT INTO \"users\" (.+)$")
	requestJSON, _ := json.Marshal(
		schema.NewUserForm{
			Username:      user.Username,
			Email:         user.Email,
			Password:      "password",
			PasswordAgain: "password",
		},
	)
	responseJSON, _ := json.Marshal(
		schema.UserOut{ID: 1, Username: user.Username, Email: user.Email},
	)
	w := request(r, http.MethodPost, "/user/", string(requestJSON))
	assertEqualCode(t, http.StatusCreated, w.Code)
	assertEqualBody(t, string(responseJSON), w.Body.String())
}

func TestUserHandler_create_duplicate(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	mockInsertQuery(
		mock,
		gorm.ErrDuplicatedKey,
		"^INSERT INTO \"users\" (.+)$",
	)
	requestJSON, _ := json.Marshal(
		schema.NewUserForm{
			Username:      user.Username,
			Email:         user.Email,
			Password:      "password",
			PasswordAgain: "password",
		},
	)
	w := request(r, http.MethodPost, "/user/", string(requestJSON))
	assertEqualCode(t, http.StatusConflict, w.Code)
}

func TestUserHandler_getById_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	// Log in.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "password"},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusOK, w.Code)

	// Get user.
	// Authentication query.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	// Request query.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	responseJSON, _ := json.Marshal(
		schema.UserOut{ID: 1, Username: user.Username, Email: user.Email},
	)
	w = request(r, http.MethodGet, "/user/1", "", w.Result().Cookies()...)
	assertEqualCode(t, http.StatusOK, w.Code)
	assertEqualBody(t, string(responseJSON), w.Body.String())
}

func TestUserHandler_getById_forbidden(t *testing.T) {
	r, _, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	w := request(r, http.MethodGet, "/user/1", "")
	assertEqualCode(t, http.StatusForbidden, w.Code)
}

func TestUserHandler_getAll_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	// Login.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "password"},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusOK, w.Code)

	// Get all users.
	// Authentication query.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	// Select query.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	responseJSON, _ := json.Marshal(
		schema.UserOut{ID: 1, Username: user.Username, Email: user.Email},
	)
	w = request(r, http.MethodGet, "/user/", "", w.Result().Cookies()...)
	assertEqualCode(t, http.StatusOK, w.Code)
	assertEqualBody(t, "["+string(responseJSON)+"]", w.Body.String())
}

func TestUserHandler_getAll_forbidden(t *testing.T) {
	r, _, err := newTestEngineWithUserHandler()
	if err != nil {
		t.Fatal(err)
	}

	w := request(r, http.MethodGet, "/user/", "")
	assertEqualCode(t, http.StatusForbidden, w.Code)
}
