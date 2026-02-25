package main

import (
	"context"
	"log"
	"net"
	"time"

	"backend/internal/config"
	"backend/internal/grpcapi"
	"backend/internal/httpapi"
	"backend/internal/shutdown"
	"backend/internal/storage"
)

func main() {
	cfg := config.Load()
	log.Println("backend up")

	ctx, cancel := context.WithCancel(context.Background())

	db := storage.OpenWithPath(cfg.DBPath)

	go httpapi.StartWithAddrCtx(ctx, db, cfg.HTTPAddr, cfg.APIKey)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}

	go grpcapi.StartCtx(ctx, lis)

	shutdown.Wait(5*time.Second, cancel)

	_ = db.Close()
	log.Println("backend down")
}
