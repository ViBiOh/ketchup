package middleware

import (
	"net/http"

	"github.com/ViBiOh/ketchup/pkg/model"
)

type Service struct {
	user model.UserService
}

func New(userService model.UserService) Service {
	return Service{
		user: userService,
	}
}

func (s Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next != nil {
			next.ServeHTTP(w, r.WithContext(s.user.StoreInContext(r.Context())))
		}
	})
}
