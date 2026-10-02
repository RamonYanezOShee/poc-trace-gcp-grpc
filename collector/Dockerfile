# Collector de OpenTelemetry de Google, con nuestra configuración incluida.
# Verifica en la documentación de Google la etiqueta (tag) vigente de la imagen.
FROM us-docker.pkg.dev/cloud-ops-agents-artifacts/google-cloud-opentelemetry-collector/otelcol-google:latest
COPY config.yaml /etc/otelcol-google/config.yaml
CMD ["--config=/etc/otelcol-google/config.yaml"]
