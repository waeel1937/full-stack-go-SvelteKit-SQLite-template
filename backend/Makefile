APP=server
CLI=grpc-check

include .env
export

.PHONY: build run grpc clean

build:
	go build -o $(APP) ./cmd/server
	go build -o $(CLI) ./cmd/grpc-check

run:
	./$(APP)

grpc:
	./$(CLI) localhost$(GRPC_ADDR)

clean:
	rm -f $(APP) $(CLI)
