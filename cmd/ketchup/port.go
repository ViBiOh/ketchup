package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/ketchup/pkg/middleware"
)

func newPort(clients clients, services services) http.Handler {
	authMux := http.NewServeMux()
	authMux.Handle("/ketchups/{id...}", services.ketchup.Ketchups())
	authMux.Handle("/", services.renderer.Handler(services.ketchup.TemplateFunc))

	mux := http.NewServeMux()
	mux.Handle("/signup", services.ketchup.Signup())
	mux.Handle("/logout", services.ketchup.Logout())
	mux.Handle("/app/", http.StripPrefix("/app", services.authMiddleware.Middleware(middleware.New(services.user).Middleware(authMux))))

	services.renderer.RegisterMux(mux, services.ketchup.PublicTemplateFunc)

	return httputils.Handler(mux, clients.health,
		clients.telemetry.Middleware("http"),
		services.owasp.Middleware,
		services.cors.Middleware,
	)
}
