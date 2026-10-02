REGION  ?= us-central1
SERVICE ?= hello-grpc

# Genera el código Go a partir del .proto (requiere protoc, protoc-gen-go y protoc-gen-go-grpc)
proto:
	protoc --proto_path=proto \
	  --go_out=gen --go_opt=paths=source_relative \
	  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	  proto/hello/v1/hello.proto
	go mod tidy

run:
	go run ./cmd/server

call:
	go run ./cmd/client -addr localhost:8080 -insecure

# Compila en Cloud Build y despliega (--use-http2 es obligatorio para gRPC; --no-cpu-throttling deja CPU para enviar trazas)
deploy:
	gcloud run deploy $(SERVICE) --source . --region $(REGION) --use-http2 --no-allow-unauthenticated \
	  --no-cpu-throttling --set-env-vars OTEL_SERVICE_NAME=$(SERVICE)

.PHONY: proto run call deploy
