// General utility functions for testing of the api package.
package api_test

import (
	"database/sql/driver"
	"fmt"
	"gin-gorm-api/config"
	"gin-gorm-api/middleware"
	"gin-gorm-api/model"
	"gin-gorm-api/provider"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testConf holds information for the testing server.
var testConf = config.Config{ //nolint:gochecknoglobals // .
	Testing: true,
	Secret: "i0LHTkqZtzzGVM8+zhQHbiPXq+ghcKpJ7wveLBppWFGe1V/m6lsp6mPwB" +
		"ndZyDUK73+iMzARrRILSmNDQz3qkg==",
}

// newAuth returns a new authentication manager and middleware for testing.
func newAuth(
	db *gorm.DB,
) (provider.UserAuthManager, gin.HandlerFunc, error) {
	mailer := provider.NewMailer(testConf)
	manager, err := provider.NewUserAuthManager(db, mailer, testConf, "user")
	if err != nil {
		return provider.UserAuthManager{}, nil, err
	}
	return manager, middleware.NewSessionMiddleware(manager), nil
}

// newMockDB returns a database session and a mock instance that can be used
// to specify the queries that should be made and the output they will return.
func newMockDB() (db *gorm.DB, mock sqlmock.Sqlmock, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("fail to initialize mock db: %w", err)
		}
	}()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	db, err = gorm.Open(
		postgres.New(postgres.Config{Conn: mockDB}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return nil, nil, err
	}
	return db, mock, nil
}

// newTestUser returns a model.User with username "user", email
// "test@email.com" and password "password".
func newTestUser(t *testing.T) model.User {
	user := model.User{ID: 1, Username: "user", Email: "test@email.com"}
	if err := user.SetPassword("password"); err != nil {
		t.Fatal(err)
	}
	return user
}

// A different function is needed for SELECTion, INSERTion and UPDATEing
// because the postgres driver manages this queries differently.

// mockSelectQuery adds expectation to mock to return instance when query
// is evaluated. instance can be an error if one is expected.
func mockSelectQuery(
	mock sqlmock.Sqlmock,
	instance any,
	query string,
) {
	if err, ok := instance.(error); ok {
		mock.ExpectQuery(query).WillReturnError(err)
		return
	}

	schema := []string{}
	values := []driver.Value{}
	val := reflect.ValueOf(instance).Elem()
	for i := range val.NumField() {
		schema = append(schema, val.Type().Field(i).Name)
		values = append(values, val.Field(i).Interface().(driver.Value))
	}

	rows := sqlmock.NewRows(schema).AddRow(values...)
	mock.ExpectQuery(query).WillReturnRows(rows)
}

// mockInsertQuery adds expectation to mock to return instance when query is
// executed as a transaction. instance can be an error if one is expected.
func mockInsertQuery(
	mock sqlmock.Sqlmock,
	instance any,
	query string,
) {
	if err, ok := instance.(error); ok {
		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnError(err)
		mock.ExpectRollback()
		return
	}

	var pkIndx int
	schema := []string{}
	values := []driver.Value{}

	val := reflect.ValueOf(instance).Elem()
	for i := range val.NumField() {
		schema = append(schema, val.Type().Field(i).Name)
		values = append(values, val.Field(i).Interface().(driver.Value))
		if val.Type().Field(i).Tag.Get("gorm") == "primarykey" {
			pkIndx = i
		}
	}

	rows := sqlmock.NewRows([]string{schema[pkIndx]}).AddRow(values[pkIndx])
	mock.ExpectBegin()
	mock.ExpectQuery(query).WillReturnRows(rows)
	mock.ExpectCommit()
}

// mockUpdateQuery adds expectation to mock to return instance when query is
// executed as a transaction. instance can be an error if one is expected. If
// instance is a record on the database then its expected to have a field
// anotated as `gorm:"primarykey"` of type (u)int(8,16,32,64).
func mockUpdateQuery(
	mock sqlmock.Sqlmock,
	instance any,
	query string,
) error {
	if err, ok := instance.(error); ok {
		mock.ExpectBegin()
		mock.ExpectExec(query).WillReturnError(err)
		mock.ExpectRollback()
		return nil
	}

	var (
		pk int64
		ok bool
	)
	val := reflect.ValueOf(instance).Elem()
	for i := range val.NumField() {
		if val.Type().Field(i).Tag.Get("gorm") != "primarykey" {
			continue
		}
		ok = true
		field := val.Field(i)
		switch field.Kind() { //nolint:exhaustive // .
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Int64:
			pk = field.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
			reflect.Float64:
			pk = int64(field.Uint())
		default:
			return fmt.Errorf(
				"primarykey of instance %v is not an integer",
				instance,
			)
		}
	}
	if !ok {
		return fmt.Errorf("could not find primarykey on instance %v", instance)
	}
	mock.ExpectBegin()
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(pk, 1))
	mock.ExpectCommit()
	return nil
}

// request returns a record of the response resulting from making a request
// with method, path, body and cookies to r.
func request(
	r *gin.Engine,
	method string,
	path string,
	body string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()

	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, strings.NewReader(body))
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}

	addCookies(req, cookies)
	r.ServeHTTP(w, req)
	return w
}

// addCookies adds cookies to req's header.
func addCookies(req *http.Request, cookies []*http.Cookie) {
	if len(cookies) == 0 {
		return
	}
	var b strings.Builder
	for _, c := range cookies {
		b.WriteString(c.String())
		b.WriteRune(',')
	}
	req.Header.Add("Cookie", b.String()[:b.Len()-1])
}

// assertEqualCode fails if expected != actual.
func assertEqualCode(t *testing.T, expected, actual int) {
	if expected == actual {
		return
	}
	t.Fatalf(
		"expected response code to be %d but got %d instead",
		expected,
		actual,
	)
}

// assertEqualBody fails if expected != actual.
func assertEqualBody(t *testing.T, expected, actual string) {
	if expected == actual {
		return
	}
	t.Fatalf(
		"expected body response to be %s but got %s instead",
		expected,
		actual,
	)
}
