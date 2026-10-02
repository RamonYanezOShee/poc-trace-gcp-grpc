# PoC telemetría: servicio gRPC "Hola mundo" en Cloud Run

Servicio Go + gRPC que el proxy llamará. Responde `Hola <nombre>`.
**Esta versión no tiene telemetría** (rama `main`). Cada enfoque de telemetría irá en su propia rama.

## Estructura

```
proto/hello/v1/hello.proto   contrato gRPC (SayHello)
cmd/server/main.go           el servicio (lo que se despliega)
cmd/client/main.go           cliente de prueba que simula al proxy (envía x-trace-id)
gen/                         código generado desde el .proto (no se versiona)
Dockerfile                   genera el código gRPC y compila; lo usa Cloud Run
Makefile                     atajos: proto, run, call, deploy
```

## Requisitos locales

Go 1.23+, `protoc`, y los plugins:

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
```

## Probar en local

```
make proto   # genera gen/ y descarga dependencias (una vez, o al cambiar el .proto)
make run     # terminal 1: levanta el servidor en :8080
make call    # terminal 2: responde "Hola mundo" y muestra el trace ID
```

El servidor imprime en el log los headers que recibe, así verás cómo llega el trace ID.

## Desplegar en Cloud Run

```
gcloud config set project TU_PROYECTO
make deploy REGION=us-central1 SERVICE=hello-grpc
```

Puntos clave para gRPC en Cloud Run:
- `--use-http2` es obligatorio (ya está en el Makefile).
- El servidor escucha en el puerto de la variable `PORT` (ya manejado).
- Con `--no-allow-unauthenticated`, el proxy debe enviar un ID token de Google en
  `authorization: Bearer ...` y su cuenta de servicio necesita el rol `roles/run.invoker`.
  Para una prueba rápida puedes usar `--allow-unauthenticated`.
- El cliente se conecta a `HOST:443` con TLS.

Probar el servicio desplegado desde tu máquina:

```
go run ./cmd/client -addr HOST_DEL_SERVICIO:443 -token "$(gcloud auth print-identity-token)"
```

## Integrarlo con tu proxy

En el proxy, genera el cliente desde el mismo `hello.proto` y llama a `SayHello`
enviando el trace ID como metadata (`x-trace-id` en este ejemplo; ajústalo al nombre que use tu proxy).

## Plan de ramas

1. `main`: versión base sin telemetría (esta).
2. `telemetria/opentelemetry`: OpenTelemetry (traces) exportando a Cloud Trace.
3. `telemetria/cloud-logging`: correlación de logs con trace ID (`X-Cloud-Trace-Context` / `traceparent`).
4. Otras variantes que se quieran comparar.

Crear el repo y la primera rama:

```
git init -b main && git add . && git commit -m "Base sin telemetría"
git switch -c telemetria/opentelemetry
```
