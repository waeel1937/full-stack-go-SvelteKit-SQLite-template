package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
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

	if err := store.InitAuth(); err != nil {
		panic(err)
	}
	if err := store.CreateUser("admin", "admin"); err == nil {
		log.Println("default user created: admin/admin")
	}

	rawBuffer := ringbuffer.New(10000)
	rawCapture := &aggregator.RawCapture{Bus: bus, Buffer: rawBuffer}
	agg := &aggregator.Aggregator{
		Window: time.Duration(cfg.Aggregator.WindowMs) * time.Millisecond,
		Bus:    bus,
	}
	persist := &aggregator.Persister{Bus: bus, Store: store}
	ruleEngine := rules.NewEngine(bus)

	status := api.NewStatusServer()
	httpServer := &api.Server{
		DB:         store.DB,
		Store:      store,
		Status:     status,
		Raw:        &api.RawServer{Buffer: rawBuffer},
		RuleEngine: ruleEngine,
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
	go httpServer.Run(":" + fmt.Sprint(cfg.Server.HTTPPort))
	go grpcSrv.Serve(lis)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			s := float64(t.UnixMilli()) / 1000.0
			bus.PublishMetric(core.MetricEvent{
				Time: t, Source: "sensor-1", Key: "temperature",
				Value: 55 + 35*math.Sin(s/30.0) + rand.Float64()*5, OK: true,
			})
			bus.PublishMetric(core.MetricEvent{
				Time: t, Source: "sensor-2", Key: "pressure",
				Value: 50 + 10*math.Cos(s/45.0) + rand.Float64()*3, OK: true,
			})
			vib := 5 + 5*math.Sin(s/10.0) + rand.Float64()*2
			if rand.Float64() < 0.02 {
				vib += 20
			}
			bus.PublishMetric(core.MetricEvent{
				Time: t, Source: "sensor-3", Key: "vibration",
				Value: vib, OK: true,
			})
			bus.PublishMetric(core.MetricEvent{
				Time: t, Source: "motor-1", Key: "rpm",
				Value: 1500 + 100*math.Sin(s/60.0) + rand.Float64()*20, OK: true,
			})
		case <-ctx.Done():
			grpcSrv.GracefulStop()
			return
		}
	}
}
