// Cliente de prueba que simula al proxy: llama al servicio y envía un trace ID
// en los headers. Sirve para probar local o contra Cloud Run.
//
// Local:      go run ./cmd/client -addr localhost:8080 -insecure
// Cloud Run:  go run ./cmd/client -addr mi-servicio-xxxx.a.run.app:443 -token "$(gcloud auth print-identity-token)"
package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	hellov1 "github.com/RamonYanezOShee/poc-trace-gcp-grpc/gen/hello/v1"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "host:puerto del servicio")
	useInsecure := flag.Bool("insecure", false, "sin TLS (solo para local)")
	token := flag.String("token", "", "ID token de Google (si el servicio no es público)")
	name := flag.String("name", "mundo", "nombre a saludar")
	flag.Parse()

	var creds credentials.TransportCredentials
	if *useInsecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{})
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("no se pudo conectar: %v", err)
	}
	defer conn.Close()

	// Trace ID de ejemplo (el proxy real ya genera el suyo).
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	traceID := hex.EncodeToString(b)

	md := metadata.Pairs("x-trace-id", traceID)
	if *token != "" {
		md.Set("authorization", "Bearer "+*token)
	}
	ctx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), md), 10*time.Second)
	defer cancel()

	resp, err := hellov1.NewHelloServiceClient(conn).SayHello(ctx, &hellov1.SayHelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("error en la llamada: %v", err)
	}
	log.Printf("trace_id=%s respuesta=%q", traceID, resp.GetMessage())
}
