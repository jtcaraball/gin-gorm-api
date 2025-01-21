package api_test

import (
	"encoding/json"
	"gin-gorm-api/api"
	"gin-gorm-api/schema"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func newTestEngineWithAuthHandler() (*gin.Engine, sqlmock.Sqlmock, error) {
	db, mock, err := newMockDB()
	if err != nil {
		return nil, nil, err
	}
	auth, sessionMW, err := newAuth(db)
	if err != nil {
		return nil, nil, err
	}
	r, err := api.NewEngine(testConf, api.NewAuthHandler(auth, sessionMW))
	if err != nil {
		return nil, nil, err
	}
	return r, mock, nil
}

func TestAuthHandler_login_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "password"},
	)
	responseJSON, _ := json.Marshal(
		schema.UserOut{ID: 1, Username: user.Username, Email: user.Email},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusOK, w.Code)
	assertEqualBody(t, string(responseJSON), w.Body.String())
}

func TestAuthHandler_login_forbidden(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.*)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "notpassword"},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusForbidden, w.Code)
}

func TestAuthHandler_logout_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
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

	// Logout.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	w = request(r, http.MethodDelete, "/auth/", "", w.Result().Cookies()...)
	assertEqualCode(t, http.StatusNoContent, w.Code)
}

func TestAuthHandler_logout_forbidden(t *testing.T) {
	r, _, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	w := request(r, http.MethodDelete, "/auth/", "")
	assertEqualCode(t, http.StatusForbidden, w.Code)
}

func TestAuthHandler_me_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
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

	// Get session user.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	responseJSON, _ := json.Marshal(
		schema.UserOut{ID: user.ID, Username: user.Username, Email: user.Email},
	)
	w = request(r, http.MethodGet, "/auth/me", "", w.Result().Cookies()...)
	assertEqualCode(t, http.StatusOK, w.Code)
	assertEqualBody(t, string(responseJSON), w.Body.String())
}

func TestAuthHandler_me_forbidden(t *testing.T) {
	r, _, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	garbageCookie := &http.Cookie{Name: "user_session", Value: "garbage"}
	w := request(r, http.MethodGet, "/auth/me", "", garbageCookie)
	assertEqualCode(t, http.StatusForbidden, w.Code)
}

func TestAuthHandler_changePassword_ok(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	// Login.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "password"},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusOK, w.Code)

	// Change password
	// Authentication query.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	// Update query.
	err = mockUpdateQuery(mock, &user, "^UPDATE \"users\" SET (.+)$")
	if err != nil {
		t.Fatal(err)
	}
	requestJSON, _ = json.Marshal(
		schema.PasswordChangeForm{
			Password:      "password1",
			PasswordAgain: "password1",
		},
	)
	w = request(
		r,
		http.MethodPost,
		"/auth/change_password",
		string(requestJSON),
		w.Result().Cookies()...,
	)
	assertEqualCode(t, http.StatusOK, w.Code)
}

func TestAuthHandler_changePassword_wrongForm(t *testing.T) {
	user := newTestUser(t)
	r, mock, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	// Login.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	requestJSON, _ := json.Marshal(
		schema.LoginForm{Username: user.Username, Password: "password"},
	)
	w := request(r, http.MethodPost, "/auth/", string(requestJSON))
	assertEqualCode(t, http.StatusOK, w.Code)

	// Change password.
	mockSelectQuery(mock, &user, "^SELECT (.+) FROM \"users\" WHERE (.+)$")
	requestJSON, _ = json.Marshal(
		schema.PasswordChangeForm{
			Password:      "password1",
			PasswordAgain: "password2",
		},
	)
	w = request(
		r,
		http.MethodPost,
		"/auth/change_password",
		string(requestJSON),
		w.Result().Cookies()...,
	)
	assertEqualCode(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_changePassword_forbidden(t *testing.T) {
	r, _, err := newTestEngineWithAuthHandler()
	if err != nil {
		t.Fatal(err)
	}

	requestJSON, _ := json.Marshal(
		schema.PasswordChangeForm{
			Password:      "password1",
			PasswordAgain: "password2",
		},
	)
	w := request(
		r,
		http.MethodPost,
		"/auth/change_password",
		string(requestJSON),
	)
	assertEqualCode(t, http.StatusForbidden, w.Code)
}
