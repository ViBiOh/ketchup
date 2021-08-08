package middleware

import (
	"net/http"

	"github.com/ViBiOh/ketchup/pkg/model"
)

// App of package
type App struct {
	userService model.UserService
}

// New creates new App from Config
func New(userService model.UserService) App {
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
