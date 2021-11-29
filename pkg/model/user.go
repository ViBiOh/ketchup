package model

import (
	"context"
	"fmt"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
)

type key int

const (
	ctxUserKey key = iota
)

// User of app
type User struct {
	Email string         `json:"email"`
	Login authModel.User `json:"login"`
	ID    uint64         `json:"id"`
}

// String implements stringer
func (u User) String() string {
	output := fmt.Sprintf("id=%d", u.ID)

	if len(u.Login.Login) != 0 {
		output += fmt.Sprintf("login=`%s`", u.Login.Login)
	} else if len(u.Email) != 0 {
		output += fmt.Sprintf("email=`%s`", u.Email)
	}

	return output
}

// IsZero check if instance is valued or not
func (u User) IsZero() bool {
	return u.ID == 0 && u.Login.ID == 0
}

// NewUser creates new User instance
func NewUser(id uint64, email string, login authModel.User) User {
	return User{
		ID:    id,
		Email: email,
		Login: login,
	}
}

// StoreUser stores given User in context
func StoreUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

// ReadUser retrieves user from context
func ReadUser(ctx context.Context) User {
	rawUser := ctx.Value(ctxUserKey)
	if rawUser == nil {
		return User{}
	}

	if user, ok := rawUser.(User); ok {
		return user
	}

	return User{}
}
