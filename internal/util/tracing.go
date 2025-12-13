package util

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() func() {
	// Default SDK TracerProvider without exporter => low-cost/safe default (acts as no-op)
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	log.Println("otel tracer initialized (no-op default)")
	return func() {
		// shutdown provider
		_ = tp.Shutdown(context.Background())
	}
}

func Tracer() trace.Tracer {
	return otel.Tracer("payment-gateway")
}
