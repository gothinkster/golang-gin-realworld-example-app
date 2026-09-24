package main

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"

	// otelc only parents spans created without a context.Context (such as
	// GORM's database/sql calls) to the active request span when the SDK trace
	// package is part of the application's own build graph.
	_ "go.opentelemetry.io/otel/sdk/trace"
)

// shutdownTelemetry flushes and stops the SDK providers injected by otelc at
// build time. It is a no-op for builds without otelc.
func shutdownTelemetry(ctx context.Context) error {
	var err error
	for _, p := range []any{otel.GetTracerProvider(), otel.GetMeterProvider(), global.GetLoggerProvider()} {
		if s, ok := p.(interface{ Shutdown(context.Context) error }); ok {
			err = errors.Join(err, s.Shutdown(ctx))
		}
	}
	return err
}
