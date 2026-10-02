package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	hellov1 "github.com/RamonYanezOShee/poc-trace-gcp-grpc/gen/hello/v1"
	"github.com/RamonYanezOShee/poc-trace-gcp-grpc/internal/telemetry"
)

var tracer = otel.Tracer("hello-grpc")

type server struct {
	hellov1.UnimplementedHelloServiceServer
}

func (s *server) SayHello(ctx context.Context, req *hellov1.SayHelloRequest) (*hellov1.SayHelloResponse, error) {
	// Span propio de ejemplo, hijo del span que crea otelgrpc para la llamada.
	ctx, span := tracer.Start(ctx, "armar-saludo")
	defer span.End()

	log.Printf("trace_id=%s", trace.SpanContextFromContext(ctx).TraceID())
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			log.Printf("header recibido: %s=%v", k, v)
		}
	}

	name := req.GetName()
	if name == "" {
		name = "mundo"
	}
	return &hellov1.SayHelloResponse{Message: "Hola " + name}, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "hello-grpc"
	}
	shutdownTelemetry, err := telemetry.Setup(ctx, serviceName)
	if err != nil {
		log.Fatalf("telemetría: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("no se pudo escuchar en el puerto %s: %v", port, err)
	}

	// otelgrpc lee el traceparent entrante y crea un span por cada llamada.
	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	hellov1.RegisterHelloServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		<-ctx.Done() // Cloud Run envía SIGTERM al apagar
		s.GracefulStop()
	}()

	log.Printf("servidor gRPC escuchando en :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("error sirviendo: %v", err)
	}

	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdownTelemetry(flushCtx); err != nil {
		log.Printf("error cerrando telemetría: %v", err)
	}
}
