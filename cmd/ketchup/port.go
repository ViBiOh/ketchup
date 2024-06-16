package main

import (
	"net/http"

	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/ketchup"
	"github.com/ViBiOh/ketchup/pkg/middleware"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

func newPort(ketchupService ketchup.Service, auth authMiddleware.Service, userService user.Service, rendererService *renderer.Service) http.Handler {
	authMux := http.NewServeMux()
	authMux.Handle("/ketchups/{id...}", ketchupService.Ketchups())
	authMux.Handle("/", rendererService.Handler(ketchupService.TemplateFunc))

	mux := http.NewServeMux()
	mux.Handle("/signup", ketchupService.Signup())
	mux.Handle("/app/{any...}", http.StripPrefix("/app", auth.Middleware(middleware.New(userService).Middleware(authMux))))

	rendererService.Register(mux, ketchupService.PublicTemplateFunc)

	return mux
}
