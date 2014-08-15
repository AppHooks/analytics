package models

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/jinzhu/gorm"
)

type User struct {
	BaseModel
	Username string
	Hash     string
}

func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"username": u.Username,
	}
}

func (u *User) Authenticate(password string) (success bool) {
	sum := sha1.Sum([]byte(password))
	return u.Hash == hex.EncodeToString(sum[:])
}

func NewUser(db *gorm.DB, username string, password string) *User {
	sum := sha1.Sum([]byte(password))
	user := User{
		Username: username,
		Hash:     hex.EncodeToString(sum[:]),
	}
	user.BaseModel = NewBaseModel(db, &user)
	return &user
}
