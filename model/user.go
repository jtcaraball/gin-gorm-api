package model

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

// User represents its namesake in the application.
type User struct {
	gorm.Model `gorm:"embedded"`
	Username   string `gorm:"unique;type:varchar(256)"`
	Email      string `gorm:"unique;type:varchar(256)"`
	Salt       []byte `json:"-" gorm:"size:8"`
	Password   []byte `json:"-" gorm:"size:32"`
}

// SetPassword sets u corresponding fields such that it can be authenticated
// using pw.
func (u *User) SetPassword(pw string) error {
	if err := u.newSalt(); err != nil {
		return err
	}
	// Following recommendation from:
	// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2
	u.Password = pbkdf2.Key([]byte(pw), u.Salt, 600000, 32, sha256.New)
	return nil
}

// CheckPassword return true if and only if pw corresponds to the string with
// which u.SetPassword was called.
func (u *User) CheckPassword(pw string) bool {
	check := pbkdf2.Key([]byte(pw), u.Salt, 600000, 32, sha256.New)
	return bytes.Equal(u.Password, check)
}

// newSalt set's u.Salt to a random 8 byte string.
func (u *User) newSalt() error {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	u.Salt = b
	return nil
}
