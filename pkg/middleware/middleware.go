package middleware

import (
	"context"
	"net/http"
)

// UserService for storing user in context
type UserService interface {
	StoreInContext(context.Context) context.Context
}

// App of package
type App struct {
	userService UserService
}

// New creates new App from Config
func New(userService UserService) App {
	return App{
		userService: userService,
	}
}

// Middleware for use with net/http
func (a App) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next != nil {
			next.ServeHTTP(w, r.WithContext(a.userService.StoreInContext(r.Context())))
		}
	})
}
