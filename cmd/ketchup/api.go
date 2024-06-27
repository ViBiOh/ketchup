package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients, err := newClients(ctx, config)
	logger.FatalfOnErr(ctx, err, "clients")

	go clients.Start()
	defer clients.Close(ctx)

	services, err := newServices(ctx, config, clients)
	logger.FatalfOnErr(ctx, err, "services")

	port := newPort(config, services)

	go services.server.Start(clients.health.EndCtx(), httputils.Handler(port, clients.health, clients.telemetry.Middleware("http"), services.owasp.Middleware, services.cors.Middleware))

	clients.health.WaitForTermination(services.server.Done())

	server.GracefulWait(services.server.Done())
}
