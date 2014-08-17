package models

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/dustin/randbo"
	"github.com/jinzhu/gorm"
	"io"
)

type User struct {
	BaseModel
	Id       int64
	Username string
	Password string
	Email    string
	Key      string
	Services []Service
}

func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"username": u.Username,
		"email":    u.Email,
	}
}

func (u *User) Authenticate(password string) (success bool) {
	sum := sha1.Sum([]byte(password))
	return u.Password == hex.EncodeToString(sum[:])
}

func GetUserFromUsername(db *gorm.DB, username string) User {
	var user User
	db.Where("username = ?", username).First(&user)
	user.BaseModel = NewBaseModel(db, &user)
	return user
}

func IsUserExists(db *gorm.DB, username string, email string) bool {
	var count int
	db.Model(User{}).Where("username = ?", username).Or("email = ?", email).Count(&count)
	return count > 0
}

func NewUser(db *gorm.DB, username string, password string, email string) *User {
	var buff bytes.Buffer
	io.CopyN(&buff, randbo.New(), 16)

	sum := sha1.Sum([]byte(password))
	user := User{
		Username: username,
		Email:    email,
		Password: hex.EncodeToString(sum[:]),
		Key:      hex.EncodeToString(buff.Bytes()),
	}
	user.BaseModel = NewBaseModel(db, &user)
	return &user
}
