module github.com/RamonYanezOShee/poc-trace-gcp-grpc

go 1.23

require (
	cloud.google.com/go/compute/metadata v0.5.2
	go.opentelemetry.io/contrib/detectors/gcp v1.31.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)
