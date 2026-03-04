FROM golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache gcc musl-dev protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
COPY go.mod go.sum ./
RUN go mod download
COPY edge-app/internal/ ./internal/
COPY edge-app/cmd/ ./cmd/
COPY edge-app/config/ ./config/
RUN protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. internal/api/grpc/pb/edge.proto
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o edge-app-bin ./cmd/edge-app

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && \
    addgroup -S edge && adduser -S edge -G edge
WORKDIR /opt/edge-app
COPY --from=build /app/edge-app-bin /opt/edge-app/edge-app
COPY --from=build /app/config /opt/edge-app/config
RUN chown -R edge:edge /opt/edge-app
USER edge
EXPOSE 8080 9090
HEALTHCHECK --interval=10s --timeout=2s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v1/status || exit 1
ENTRYPOINT ["/opt/edge-app/edge-app"]
