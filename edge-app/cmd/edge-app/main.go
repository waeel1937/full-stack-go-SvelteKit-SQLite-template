package main

import (
	"net"
	"time"

	"edge-app/internal/aggregator"
	grpcapi "edge-app/internal/api/grpc"
	pb "edge-app/internal/api/grpc/pb"
	"edge-app/internal/api"
	"edge-app/internal/core"
	"edge-app/internal/rules"
	"edge-app/internal/storage"

	"google.golang.org/grpc"
)

func main() {
	bus := core.NewBus()

	store, err := storage.Open("edge.db")
	if err != nil {
		panic(err)
	}

	agg := &aggregator.Aggregator{
		Window: time.Second,
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

	httpServer := &api.Server{
		DB: store.DB,
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterEdgeServiceServer(grpcSrv, &grpcapi.Server{DB: store.DB})

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	go agg.Run()
	go persist.Run()
	go ruleEngine.Run()
	go httpServer.Run(":8080")
	go grpcSrv.Serve(lis)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for t := range ticker.C {
		bus.Metrics <- core.MetricEvent{
			Time:   t,
			Source: "demo",
			Key:    "temperature",
			Value:  float64(t.Unix() % 120),
			OK:     true,
		}
	}
}
