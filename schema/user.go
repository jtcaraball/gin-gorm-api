package schema

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// ============================================== //
//                    INPUT                       //
// ============================================== //

// NewUserForm contains the necessary information to create a new user.
type NewUserForm struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
}

// Validate f's schema.
func (f NewUserForm) Validate() (Errors, error) {
	err := validation.ValidateStruct(
		&f,
		validation.Field(
			&f.Username,
			validation.Required,
			validation.Length(4, 16),
			is.Alphanumeric,
		),
		validation.Field(
			&f.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(
			&f.Password,
			validation.Required,
			validation.Length(8, 256),
		),
		validation.Field(
			&f.PasswordAgain,
			validation.Required,
			validation.Length(8, 256),
			validation.By(matchingFieldsRule(f.Password, "password")),
		),
	)
	return errToErrors(err)
}

// LoginForm contains the information required to authenticate a user.
type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate f's schema.
func (f LoginForm) Validate() (Errors, error) {
	err := validation.ValidateStruct(
		&f,
		validation.Field(
			&f.Username,
			validation.Required,
			is.Alphanumeric,
		),
		validation.Field(
			&f.Password,
			validation.Required,
			validation.Length(8, 256),
		),
	)
	return errToErrors(err)
}

// PasswordResetRequestForm contains the information required to send a
// password reset token.
type PasswordResetRequestForm struct {
	Email string `json:"email"`
}

// Validate f's schema.
func (f PasswordResetRequestForm) Validate() (Errors, error) {
	err := validation.ValidateStruct(
		&f,
		validation.Field(
			&f.Email,
			validation.Required,
			is.Email,
		),
	)
	return errToErrors(err)
}

// PasswordResetForm contains the information required to change the
// password of an unauthenticated user.
type PasswordResetForm struct {
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
	Token         string `json:"token"`
}

// Validate f's schema.
func (f PasswordResetForm) Validate() (Errors, error) {
	err := validation.ValidateStruct(
		&f,
		validation.Field(
			&f.Password,
			validation.Required,
			validation.Length(8, 256),
		),
		validation.Field(
			&f.PasswordAgain,
			validation.Required,
			validation.Length(8, 256),
			validation.By(matchingFieldsRule(f.Password, "password")),
		),
		validation.Field(
			&f.Token,
			validation.Required,
		),
	)
	return errToErrors(err)
}

// PasswordChangeForm contains the information required to change the
// password of an authenticated user.
type PasswordChangeForm struct {
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
}

// Validate f's schema.
func (f PasswordChangeForm) Validate() (Errors, error) {
	err := validation.ValidateStruct(
		&f,
		validation.Field(
			&f.Password,
			validation.Required,
			validation.Length(8, 256),
		),
		validation.Field(
			&f.PasswordAgain,
			validation.Required,
			validation.Length(8, 256),
			validation.By(matchingFieldsRule(f.Password, "Password")),
		),
	)
	return errToErrors(err)
}

// ============================================== //
//                    OUTPUT                      //
// ============================================== //

// UserOut contains information about a user.
type UserOut struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
