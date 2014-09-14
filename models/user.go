package models

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"

	"github.com/dustin/randbo"
	"github.com/jinzhu/gorm"
	"github.com/llun/analytics/validators"
)

type User struct {
	BaseModel `sql:"-"`
	Id        int64
	Email     string
	Password  string
	Key       string
	Services  []Service
}

func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"email": u.Email,
	}
}

func (u *User) Authenticate(password string) bool {
	sum := sha1.Sum([]byte(password))
	return u.Password == hex.EncodeToString(sum[:])
}

func (u *User) UpdatePassword(password string, confirm string) bool {
	if password != confirm {
		return false
	} else if password == "" && confirm == "" {
		return true
	}

	sum := sha1.Sum([]byte(password))
	u.Password = hex.EncodeToString(sum[:])
	return true
}

func getUserFromField(db *gorm.DB, field string, value interface{}) *User {
	var (
		user  User
		total int
	)
	db.Where(field+" = ?", value).First(&user).Count(&total)

	if total == 0 {
		return nil
	}
	user.BaseModel = NewBaseModel(db, &user)
	return &user
}

func GetUserFromEmail(db *gorm.DB, email string) *User {
	return getUserFromField(db, "email", email)
}

func GetUserFromKey(db *gorm.DB, key string) *User {
	return getUserFromField(db, "key", key)
}

func GetUserFromId(db *gorm.DB, id int64) *User {
	return getUserFromField(db, "id", id)
}

func IsUserExists(db *gorm.DB, email string) bool {
	var count int
	db.Model(User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func NewUser(db *gorm.DB, email string, password string) (*User, error) {
	var buff bytes.Buffer

	if !validators.IsEmail(email) {
		return nil, errors.New("Invalid Email")
	}

	// Generate Key Hash
	io.CopyN(&buff, randbo.New(), 16)

	// Create password hash from sha1 algorithm
	sum := sha1.Sum([]byte(password))
	user := User{
		Email:    email,
		Password: hex.EncodeToString(sum[:]),
		Key:      hex.EncodeToString(buff.Bytes()),
	}

	// Extension - Golang doesn't inheritance.
	user.BaseModel = NewBaseModel(db, &user)
	return &user, nil
}
