package main

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

type clients struct {
	telemetry *telemetry.Service
	pprof     *pprof.Service
	health    *health.Service

	redis redis.Client
	db    db.Service
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

	service, version, env := output.telemetry.GetServiceVersionAndEnv()
	output.pprof = pprof.New(config.pprof, service, version, env)

	output.db, err = db.New(ctx, config.db, output.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("database: %w", err)
	}

	output.redis, err = redis.New(ctx, config.redis, output.telemetry.MeterProvider(), output.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("redis: %w", err)
	}

	output.health = health.New(ctx, config.health, output.db.Ping)

	return output, nil
}

func (c clients) Start() {
	go c.pprof.Start(c.health.DoneCtx())
}

func (c clients) Close(ctx context.Context) {
	c.db.Close()
	c.redis.Close(ctx)
	c.telemetry.Close(ctx)
}
