package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gin-gorm-api/model"
	"gin-gorm-api/schema"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 255 tokens is plenty.
type tokenType uint8

const (
	_ tokenType = iota
	sessionToken
	resetToken
)

// An authToken is a signed string that identifies a user and a time frame for
// it to be used.
type authToken struct {
	Info             authTokenInfo `json:"info"`
	VerificationCode string        `json:"verification_code"`
}

// authTokenInfo holds the user and time frame information of an authToken.
type authTokenInfo struct {
	UserID    uint      `json:"user_id"`
	Type      tokenType `json:"type"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// A UserAuthManager can perform basic authentication tasks based on
// model.User. It uses HMAC-SHA256 for token signing.
type UserAuthManager struct {
	db      *gorm.DB
	msm     Mailer
	secret  []byte
	UserKey string
}

// NewUserAuthManager returns a UserAuthManager.
func NewUserAuthManager(
	db *gorm.DB,
	secret, userKey string,
	msm Mailer,
) (manager UserAuthManager, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create manager: %w", err)
		}
	}()

	secretB, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return UserAuthManager{}, err
	}
	if len(secretB) != 64 {
		return UserAuthManager{}, ErrInvalidSecretSize
	}
	manager = UserAuthManager{
		db:      db,
		msm:     msm,
		secret:  secretB,
		UserKey: userKey,
	}
	return manager, nil
}

// Authenticate the credentials in form and returns their corresponding user.
func (m UserAuthManager) Authenticate(
	form schema.LoginForm,
	c *gin.Context,
) (user model.User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to authenticate: %w", err)
		}
	}()

	r := m.db.WithContext(c).First(&user, "username = ?", form.Username)
	if r.Error != nil {
		return model.User{}, r.Error
	}
	if !user.CheckPassword(form.Password) {
		return model.User{}, ErrInvalidCredentials
	}
	return user, nil
}

// RegisterSession generates an authentication token for user and calls
// c.SetCookie with it.
func (m UserAuthManager) RegisterSession(
	user model.User,
	c *gin.Context,
) error {
	seconds := 3600
	token := m.signedToken(
		user.ID,
		sessionToken,
		time.Now(),
		time.Duration(seconds)*time.Second,
	)
	tokenB, err := json.Marshal(&token)
	if err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}
	tokenS := base64.StdEncoding.EncodeToString(tokenB)
	c.SetCookie("user_session", tokenS, seconds, "/", "", false, true)
	return nil
}

// RetrieveSession call c.Cookie to obtain a session's authentication token and
// if a valid one is found returns the corresponding user.
func (m UserAuthManager) RetrieveSession(
	c *gin.Context,
) (user model.User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to retrieve session: %w", err)
		}
	}()

	s, err := c.Cookie("user_session")
	if err != nil {
		return user, err
	}
	token, err := m.parseToken(s)
	if err != nil {
		return user, err
	}
	if r := m.db.WithContext(c.Request.Context()).First(
		&user,
		token.Info.UserID,
	); r.Error != nil {
		return user, r.Error
	}
	return user, nil
}

// RemoveSession sets an empty user session cookie.
func (m UserAuthManager) RemoveSession(c *gin.Context) {
	c.SetCookie("user_session", "", -1, "/", "", false, true)
}

// RequestPasswordReset generates a password reset token and calls m.msm.Send
// with it.
func (m UserAuthManager) RequestPasswordReset(
	form schema.PasswordResetRequestForm,
	c *gin.Context,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("could not reset request: %w", err)
		}
	}()

	var user model.User
	if r := m.db.WithContext(c.Request.Context()).First(
		&user,
		"email = ?",
		form.Email,
	); r.Error != nil {
		return r.Error
	}
	token := m.signedToken(user.ID, resetToken, time.Now(), 10*time.Minute)
	tokenB, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return m.msm.Send(
		c.Request.Context(),
		form.Email,
		"Password reset code",
		base64.StdEncoding.EncodeToString(tokenB),
	)
}

// ResetPassword validates a password reset token and if it is valid changes
// the corresponding user's password according to the information in form.
func (m UserAuthManager) ResetPassword(
	form schema.PasswordResetForm,
	c *gin.Context,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to reset password: %w", err)
		}
	}()

	token, err := m.parseToken(form.Token)
	if err != nil {
		return err
	}
	var user model.User
	if r := m.db.WithContext(c.Request.Context()).First(
		&user,
		token.Info.UserID,
	); r.Error != nil {
		return r.Error
	}
	if user.UpdatedAt.After(token.Info.IssuedAt) {
		return ErrTokenExpired
	}
	if err = user.SetPassword(form.Password); err != nil {
		return err
	}
	if r := m.db.WithContext(c.Request.Context()).Model(&user).Update(
		"password",
		user.Password,
	); r.Error != nil {
		return r.Error
	}
	return nil
}

// SetPassword changes the user's password to match the one in form.
func (m UserAuthManager) SetPassword(
	user model.User,
	form schema.PasswordChangeForm,
	c *gin.Context,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to set password: %w", err)
		}
	}()

	if err = user.SetPassword(form.Password); err != nil {
		return err
	}
	if r := m.db.WithContext(c.Request.Context()).Model(&user).Update(
		"password",
		user.Password,
	); r.Error != nil {
		return r.Error
	}
	return nil
}

// parseToken returns the token encoded in s provided that s is a valid
// encoding.
func (m UserAuthManager) parseToken(s string) (authToken, error) {
	var token authToken
	decToken, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return token, ErrInvalidToken
	}
	if err = json.Unmarshal(decToken, &token); err != nil {
		return token, fmt.Errorf("could not parse token: %w", err)
	}
	if !m.validToken(token) {
		return token, ErrInvalidToken
	}
	return token, nil
}

// validToken returns true if and only if its verification code is a valid
// signature of its information.
func (m UserAuthManager) validToken(token authToken) bool {
	verToken := m.signedToken(
		token.Info.UserID,
		token.Info.Type,
		token.Info.IssuedAt,
		token.Info.ExpiresAt.Sub(token.Info.IssuedAt),
	)
	code, err := base64.StdEncoding.DecodeString(token.VerificationCode)
	if err != nil {
		return false // This could be a bad token being given so not a problem.
	}
	verCode, err := base64.StdEncoding.DecodeString(verToken.VerificationCode)
	if err != nil {
		panic(err) // We generated this code so it is our problem.
	}
	if !hmac.Equal(code, verCode) {
		return false
	}
	if time.Now().After(token.Info.ExpiresAt) {
		return false
	}
	return true
}

// signedToken returns a valid authentication token with the provided
// information.
func (m UserAuthManager) signedToken(
	id uint,
	t tokenType,
	issued time.Time,
	duration time.Duration,
) authToken {
	tokenInfo := authTokenInfo{id, t, issued, issued.Add(duration)}
	info, err := json.Marshal(&tokenInfo)
	if err != nil {
		panic(err)
	}
	hmac := hmac.New(sha256.New, m.secret)
	hmac.Write(info)
	return authToken{
		tokenInfo,
		base64.StdEncoding.EncodeToString(hmac.Sum(nil)),
	}
}
