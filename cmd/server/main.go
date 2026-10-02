package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	hellov1 "github.com/RamonYanezOShee/poc-trace-gcp-grpc/gen/hello/v1"
)

type server struct {
	hellov1.UnimplementedHelloServiceServer
}

func (s *server) SayHello(ctx context.Context, req *hellov1.SayHelloRequest) (*hellov1.SayHelloResponse, error) {
	// Muestra los headers (metadata) que llegan desde el proxy.
	// Útil para ver cómo viaja el trace ID antes de agregar telemetría.
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
	// Cloud Run indica el puerto mediante la variable PORT.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("no se pudo escuchar en el puerto %s: %v", port, err)
	}

	s := grpc.NewServer()
	hellov1.RegisterHelloServiceServer(s, &server{})
	reflection.Register(s) // permite probar con grpcurl sin el .proto

	log.Printf("servidor gRPC escuchando en :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("error sirviendo: %v", err)
	}
}
