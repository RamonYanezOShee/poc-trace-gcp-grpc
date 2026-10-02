// Package telemetry configura OpenTelemetry para enviar trazas por OTLP/gRPC
// al Collector que corre como sidecar en el mismo servicio de Cloud Run
// (localhost:4317). El Collector se encarga de autenticarse y enviar a Google.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const defaultEndpoint = "localhost:4317"

// Setup deja listo el tracer global y devuelve una función para vaciar y cerrar
// el exportador (hay que llamarla al apagar el servicio).
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	// Es tráfico interno del mismo contenedor de red, por eso sin TLS.
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("exportador OTLP: %w", err)
	}

	// Los datos de Cloud Run y el proyecto los agrega el Collector (resourcedetection).
	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	// W3C traceparent: es el formato con el que el proxy debe enviar el trace ID.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
