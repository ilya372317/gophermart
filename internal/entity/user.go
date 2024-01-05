package entity

import (
	"crypto/sha256"
	"encoding/hex"
)

type User struct {
	Login    string `db:"login"`
	Password string `db:"password"`
	ID       uint   `db:"id"`
}

func (u *User) SetPassword(pass string) {
	hash := sha256.Sum256([]byte(pass))
	u.Password = hex.EncodeToString(hash[:])
}

func (u *User) GetPasswordHash() string {
	return u.Password
}

func (u *User) CheckPassword(pass string) bool {
	hash := sha256.Sum256([]byte(pass))
	return u.Password == hex.EncodeToString(hash[:])
}
