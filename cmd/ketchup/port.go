package main

import (
	"net/http"

	"github.com/ViBiOh/ketchup/pkg/middleware"
)

func newPort(config configuration, services services) http.Handler {
	authMux := http.NewServeMux()
	authMux.Handle("/ketchups/{id...}", services.ketchup.Ketchups())
	authMux.Handle("/", services.renderer.Handler(services.ketchup.TemplateFunc))

	mux := http.NewServeMux()
	mux.Handle("/signup", services.ketchup.Signup())
	mux.Handle("/app/", http.StripPrefix("/app", services.authMiddleware.Middleware(middleware.New(services.user).Middleware(authMux))))

	mux.Handle(config.renderer.PathPrefix+"/", http.StripPrefix(
		config.renderer.PathPrefix,
		services.renderer.NewServeMux(services.ketchup.PublicTemplateFunc),
	))

	return mux
}
