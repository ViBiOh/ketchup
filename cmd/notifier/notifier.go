package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

func main() {
	config := newConfig()

	ctx := context.Background()

	clients, err := newClients(ctx, config)
	logger.FatalfOnErr(ctx, err, "clients")

	defer clients.Close(ctx)

	services, err := newServices(ctx, config, clients)
	logger.FatalfOnErr(ctx, err, "services")

	defer services.Close(ctx)

	slog.InfoContext(ctx, "Starting notifier...")

	ctx, end := telemetry.StartSpan(ctx, clients.telemetry.TracerProvider().Tracer("notifier"), "notifier")
	defer end(&err)

	if err = services.notifier.Notify(ctx); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "notify", slog.Any("error", err))
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Notifier ended!")
}
