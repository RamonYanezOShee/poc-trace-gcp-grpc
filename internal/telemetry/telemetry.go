// Package telemetry configura OpenTelemetry para enviar trazas directamente
// a la API de Telemetría de Google (telemetry.googleapis.com) vía OTLP/gRPC.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const endpoint = "telemetry.googleapis.com:443"

// Setup deja listo el tracer global y devuelve una función para vaciar y cerrar
// el exportador (hay que llamarla al apagar el servicio).
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		id, err := metadata.ProjectIDWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("no se pudo obtener el project ID (define GOOGLE_CLOUD_PROJECT): %w", err)
		}
		projectID = id
	}

	// Credenciales de la cuenta de servicio (en Cloud Run) o de gcloud (en local).
	creds, err := oauth.NewApplicationDefault(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("credenciales de Google: %w", err)
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithDialOption(grpc.WithPerRPCCredentials(creds)),
		otlptracegrpc.WithHeaders(map[string]string{"x-goog-user-project": projectID}),
	)
	if err != nil {
		return nil, fmt.Errorf("exportador OTLP: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()), // datos de Cloud Run (servicio, revisión, región)
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("gcp.project_id", projectID), // requerido por la API de Telemetría
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		// Envía rápido: en Cloud Run la CPU puede reducirse entre peticiones.
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
