package model

import (
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
)

var (
	// NoneUser is an undefined user
	NoneUser = User{}
)

// User of app
type User struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Login authModel.User
}
