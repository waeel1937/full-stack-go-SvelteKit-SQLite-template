package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"

	"edge-app/internal/aggregator"
	grpcapi "edge-app/internal/api/grpc"
	pb "edge-app/internal/api/grpc/pb"
	"edge-app/internal/api"
	"edge-app/internal/config"
	"edge-app/internal/connector"
	"edge-app/internal/core"
	"edge-app/internal/logging"
	"edge-app/internal/metrics"
	"edge-app/internal/rules"
	"edge-app/internal/storage"
	"edge-app/internal/storage/ringbuffer"

	"google.golang.org/grpc"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

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
	rawCapture := &aggregator.RawCapture{Bus: bus, Buffer: rawBuffer}
	agg := &aggregator.Aggregator{
		Window: time.Duration(cfg.Aggregator.WindowMs) * time.Millisecond,
		Bus:    bus,
	}
	persist := &aggregator.Persister{Bus: bus, Store: store}
	ruleEngine := rules.NewEngine(bus)

	httpServer := &api.Server{
		DB:          store.DB,
		Status:      api.NewStatusServer(),
		Raw:         &api.RawServer{Buffer: rawBuffer},
		RuleEngine:  ruleEngine,
		KeycloakURL: env("KEYCLOAK_URL", "http://localhost:8180"),
		Realm:       env("KEYCLOAK_REALM", "edge"),
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

	connector.StartAll(cfg.Connectors, bus)

	hasConnector := cfg.Connectors.Modbus.Enabled || cfg.Connectors.OPCUA.Enabled
	if !hasConnector {
		logging.Logger.Println("no connectors enabled, starting demo producer")
		go demo(ctx, bus)
	}

	<-ctx.Done()
	grpcSrv.GracefulStop()
}

func demo(ctx context.Context, bus *core.Bus) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-tick.C:
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
		}
	}
}
