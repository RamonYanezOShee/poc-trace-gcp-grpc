// Cliente de prueba que simula al proxy: genera un trace ID y lo envía en el
// header W3C "traceparent". Así el servicio continúa esa misma traza.
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
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	hellov1 "github.com/RamonYanezOShee/poc-trace-gcp-grpc/gen/hello/v1"
)

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

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

	// Formato W3C: 00-<trace-id 32 hex>-<span-id 16 hex>-<flags>; 01 = muestreado.
	traceID := randomHex(16)
	traceparent := fmt.Sprintf("00-%s-%s-01", traceID, randomHex(8))

	md := metadata.Pairs("traceparent", traceparent)
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
