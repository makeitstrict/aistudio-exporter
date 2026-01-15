FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY aistudio-exporter /usr/bin/aistudio-exporter

ENTRYPOINT ["/usr/bin/aistudio-exporter"]
