package main

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

type clients struct {
	telemetry *telemetry.Service

	db db.Service
}

func newClients(ctx context.Context, config configuration) (clients, error) {
	var output clients
	var err error

	logger.Init(ctx, config.logger)

	output.telemetry, err = telemetry.New(ctx, config.telemetry)
	if err != nil {
		return output, fmt.Errorf("telemetry: %w", err)
	}

	logger.AddOpenTelemetryToDefaultLogger(output.telemetry)
	request.AddOpenTelemetryToDefaultClient(output.telemetry.MeterProvider(), output.telemetry.TracerProvider())

	output.db, err = db.New(ctx, config.db, output.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("database: %w", err)
	}

	return output, nil
}

func (c clients) Close(ctx context.Context) {
	c.db.Close()
	c.telemetry.Close(ctx)
}
