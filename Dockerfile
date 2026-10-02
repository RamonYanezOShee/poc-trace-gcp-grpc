# Etapa 1: genera el código gRPC desde el .proto y compila el servidor
FROM golang:1.23 AS build

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /src
COPY . .
RUN mkdir -p gen && protoc --proto_path=proto \
      --go_out=gen --go_opt=paths=source_relative \
      --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
      proto/hello/v1/hello.proto
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/server ./cmd/server

# Etapa 2: imagen final mínima
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/server /server
ENTRYPOINT ["/server"]
