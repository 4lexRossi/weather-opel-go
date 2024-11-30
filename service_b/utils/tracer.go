package utils

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer initializes OTEL with a Zipkin exporter
func InitTracer(serviceName string) *sdktrace.TracerProvider {
	// Zipkin endpoint
	zipkinEndpoint := "http://zipkin:9411/api/v2/spans"

	// Create Zipkin exporter
	exporter, err := zipkin.New(
		zipkinEndpoint,
	)
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}

	// Create a trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", serviceName),
		)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Shutdown trace provider
	if err := tp.Shutdown(context.Background()); err != nil {
		log.Fatalf("Failed to shutdown tracer provider: %v", err)
	}

	return tp
}
