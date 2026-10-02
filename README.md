# PoC telemetría: servicio gRPC "Hola mundo" en Cloud Run

Servicio Go + gRPC que el proxy llamará. Responde `Hola <nombre>`.
**Rama `telemetria/otel-sidecar`:** la app envía trazas con OpenTelemetry a un Collector que corre como sidecar en el mismo servicio de Cloud Run, y el Collector las envía a Google. La base sin telemetría está en `main` y la variante sin sidecar en `telemetria/otel-directo`.

## Estructura

```
proto/hello/v1/hello.proto   contrato gRPC (SayHello)
cmd/server/main.go           el servicio (lo que se despliega)
cmd/client/main.go           cliente de prueba que simula al proxy (envía traceparent)
gen/                         código generado desde el .proto (no se versiona)
Dockerfile                   genera el código gRPC y compila la app
collector/                   imagen y configuración del Collector (sidecar)
service.yaml                 servicio de Cloud Run con 2 contenedores: app + collector
Makefile                     atajos: proto, run, call, repo, build, deploy
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

## Desplegar en Cloud Run (app + sidecar)

Configuración en Google Cloud (una vez):

```
gcloud config set project TU_PROYECTO
gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com cloudtrace.googleapis.com
# cuenta de servicio con la que corre Cloud Run (por defecto la de Compute):
gcloud projects add-iam-policy-binding TU_PROYECTO \
  --member=serviceAccount:CUENTA_DE_SERVICIO --role=roles/cloudtrace.agent
make repo     # crea el repositorio de imágenes en Artifact Registry
```

Cada despliegue:

```
make deploy REGION=us-central1 SERVICE=hello-grpc
# opcional: PROJECT=mi-proyecto SERVICE_ACCOUNT=cuenta@mi-proyecto.iam.gserviceaccount.com
```

`make deploy` construye las dos imágenes (app y Collector) y despliega `service.yaml`.

Cómo funciona:
- La app (contenedor `app`) recibe el tráfico gRPC. `otelgrpc` crea un span por llamada y continúa el `traceparent` del proxy.
- La app envía los spans por OTLP a `localhost:4317`, donde escucha el Collector (contenedor `collector`).
- El Collector (`collector/config.yaml`) agrupa, agrega datos de Cloud Run y exporta a Cloud Trace con la cuenta de servicio.
- `service.yaml` activa HTTP/2 (puerto `h2c`), CPU siempre asignada y que la app arranque después del Collector.

Con el servicio no público, el proxy debe enviar un ID token de Google (`authorization: Bearer ...`) y su cuenta
necesita `roles/run.invoker`. El cliente se conecta a `HOST:443` con TLS:

```
go run ./cmd/client -addr HOST_DEL_SERVICIO:443 -token "$(gcloud auth print-identity-token)"
```

Ver las trazas: consola de Google Cloud, Trace Explorer, filtrando por el servicio `hello-grpc`.

En local no hay Collector: `make run` funciona, pero la app mostrará errores de exportación de trazas en el log (no afectan las respuestas).

## Integrarlo con tu proxy

En el proxy, genera el cliente desde el mismo `hello.proto` y llama a `SayHello`
enviando el trace ID en el header gRPC `traceparent` (formato W3C `00-<32 hex>-<16 hex>-01`). Si el proxy usa otro formato, hay que convertirlo.

## Ramas

- `main`: base sin telemetría.
- `telemetria/otel-directo`: la app envía a `telemetry.googleapis.com` sin sidecar.
- `telemetria/otel-sidecar`: la app envía a un Collector sidecar (esta).
