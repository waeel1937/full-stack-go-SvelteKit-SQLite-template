FROM golang:1.22-alpine AS build

WORKDIR /app
COPY . .

RUN apk add --no-cache gcc musl-dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o edge-app ./cmd/edge-app

FROM alpine:3.19

RUN addgroup -S edge && adduser -S edge -G edge
WORKDIR /opt/edge-app

COPY --from=build /app/edge-app /opt/edge-app/edge-app
COPY --from=build /app/config /opt/edge-app/config

RUN chown -R edge:edge /opt/edge-app
USER edge

EXPOSE 8080 9090

HEALTHCHECK --interval=10s --timeout=2s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v1/status || exit 1

ENTRYPOINT ["/opt/edge-app/edge-app"]
