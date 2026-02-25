FROM golang:1.24-alpine AS build

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN go build -o server ./cmd/server
RUN go build -o grpc-check ./cmd/grpc-check

FROM alpine:3.19

RUN apk add --no-cache sqlite-libs

WORKDIR /app
COPY --from=build /app/server /app/server
COPY --from=build /app/grpc-check /app/grpc-check
COPY .env /app/.env

EXPOSE 8081 9091

STOPSIGNAL SIGTERM
CMD ["./server"]
