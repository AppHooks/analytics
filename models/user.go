package models

import (
	"crypto/sha1"
	"encoding/hex"
)

type User struct {
	Username string
	Hash     string
}

func (u *User) Authenticate(password string) (success bool) {
	sum := sha1.Sum([]byte(password))
	return u.Hash == hex.EncodeToString(sum[:])
}
