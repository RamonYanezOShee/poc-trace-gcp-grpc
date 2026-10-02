REGION          ?= us-central1
SERVICE         ?= hello-grpc
# Debes definir estas dos al desplegar: make deploy PROJECT=mi-proyecto SERVICE_ACCOUNT=...
PROJECT         ?= $(shell gcloud config get-value project 2>/dev/null)
SERVICE_ACCOUNT ?= $(shell gcloud projects describe $(PROJECT) --format='value(projectNumber)')-compute@developer.gserviceaccount.com
REPO            ?= poc-trace

IMAGE_APP       = $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/$(SERVICE):latest
IMAGE_COLLECTOR = $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/otel-collector:latest

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

# Crea el repositorio de imágenes (una sola vez)
repo:
	gcloud artifacts repositories create $(REPO) --repository-format=docker --location=$(REGION)

# Construye y sube las dos imágenes con Cloud Build
build:
	gcloud builds submit --tag $(IMAGE_APP) .
	gcloud builds submit --tag $(IMAGE_COLLECTOR) collector/

# Despliega los dos contenedores (app + sidecar) en un solo servicio de Cloud Run
deploy: build
	SERVICE=$(SERVICE) SERVICE_ACCOUNT=$(SERVICE_ACCOUNT) IMAGE_APP=$(IMAGE_APP) IMAGE_COLLECTOR=$(IMAGE_COLLECTOR) \
	  envsubst < service.yaml > /tmp/service.rendered.yaml
	gcloud run services replace /tmp/service.rendered.yaml --region $(REGION)

.PHONY: proto run call repo build deploy
