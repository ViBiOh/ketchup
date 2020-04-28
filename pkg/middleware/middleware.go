package middleware

import (
	"net/http"

	"github.com/ViBiOh/ketchup/pkg/service/user"
)

// App of package
type App interface {
	Middleware(next http.Handler) http.Handler
}

type app struct {
	userService user.App
}

// New creates new App from Config
func New(userService user.App) App {
	return app{
		userService: userService,
	}
}

func (a app) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next != nil {
			next.ServeHTTP(w, r.WithContext(a.userService.StoreInContext(r.Context())))
		}
	})
}
