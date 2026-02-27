package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"edge-app/internal/aggregator"
	grpcapi "edge-app/internal/api/grpc"
	pb "edge-app/internal/api/grpc/pb"
	"edge-app/internal/api"
	"edge-app/internal/config"
	"edge-app/internal/core"
	"edge-app/internal/logging"
	"edge-app/internal/metrics"
	"edge-app/internal/rules"
	"edge-app/internal/storage"
	"edge-app/internal/storage/ringbuffer"
	"edge-app/internal/sync"

	"google.golang.org/grpc"
)

func main() {
	logging.Init()
	metrics.Init()

	cfg, err := config.Load("config/app.yaml")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go core.WaitForShutdown(cancel)

	bus := core.NewBus()

	store, err := storage.Open(cfg.Database.Path)
	if err != nil {
		panic(err)
	}

	rawBuffer := ringbuffer.New(10000)

	rawCapture := &aggregator.RawCapture{
		In:     bus.Metrics,
		Buffer: rawBuffer,
	}

	agg := &aggregator.Aggregator{
		Window: time.Duration(cfg.Aggregator.WindowMs) * time.Millisecond,
		In:     bus.Metrics,
		Out:    bus.Aggregates,
	}

	persist := &aggregator.Persister{
		In:    bus.Aggregates,
		Store: store,
	}

	ruleEngine := &rules.Engine{
		In: bus.Aggregates,
	}

	status := api.NewStatusServer()

	httpServer := &api.Server{
		DB:     store.DB,
		Status: status,
		Raw:    &api.RawServer{Buffer: rawBuffer},
	}

	cloudSync := &sync.CloudSync{
		DB:       store.DB,
		Endpoint: "https://example.com/edge-sync",
		Interval: 60 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterEdgeServiceServer(grpcSrv, &grpcapi.Server{DB: store.DB})

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(cfg.Server.GRPCPort))
	if err != nil {
		panic(err)
	}

	go rawCapture.Run()
	go agg.Run()
	go persist.Run()
	go ruleEngine.Run()
	go cloudSync.Run()
	go httpServer.Run(":"+fmt.Sprint(cfg.Server.HTTPPort))
	go grpcSrv.Serve(lis)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			bus.Metrics <- core.MetricEvent{
				Time:   t,
				Source: "demo",
				Key:    "temperature",
				Value:  float64(t.Unix() % 120),
				OK:     true,
			}
		case <-ctx.Done():
			grpcSrv.GracefulStop()
			return
		}
	}
}
