package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
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
	cloudsync "edge-app/internal/sync"

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
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go core.WaitForShutdown(cancel)

	bus := core.NewBus()

	store, err := storage.Open(cfg.Database.Path)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	defer store.Close()

	ruleEngine, err := rules.NewEngine(bus, store)
	if err != nil {
		log.Fatalf("rules: %v", err)
	}

	rawBuffer := ringbuffer.New(10000)

	rawCapture := &aggregator.RawCapture{Bus: bus, Buffer: rawBuffer}
	agg := &aggregator.Aggregator{
		Window: time.Duration(cfg.Aggregator.WindowMs) * time.Millisecond,
		Bus:    bus,
	}
	persist := &aggregator.Persister{Bus: bus, Store: store}

	httpServer := &api.Server{
		DB:          store.DB,
		Bus:         bus,
		Status:      api.NewStatusServer(),
		Raw:         &api.RawServer{Buffer: rawBuffer},
		RuleEngine:  ruleEngine,
		KeycloakURL: env("KEYCLOAK_URL", "http://localhost:8180"),
		Realm:       env("KEYCLOAK_REALM", "edge"),
		CORSOrigin:  env("CORS_ORIGIN", ""),
	}

	// gRPC
	grpcSrv := grpc.NewServer()
	pb.RegisterEdgeServiceServer(grpcSrv, &grpcapi.Server{DB: store.DB, Bus: bus})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("gRPC listen: %v", err)
	}

	// HTTP — Build returns a fully configured *http.Server
	httpSrv := httpServer.Build(ctx, fmt.Sprintf(":%d", cfg.Server.HTTPPort))

	// Cloud sync (optional — only runs if endpoint is configured)
	if endpoint := env("CLOUD_SYNC_ENDPOINT", ""); endpoint != "" {
		cs := &cloudsync.CloudSync{
			Store:    store,
			Endpoint: endpoint,
			Interval: 30 * time.Second,
		}
		go cs.Run(ctx)
	}

	go rawCapture.Run(ctx)
	go agg.Run(ctx)
	go persist.Run(ctx)
	go ruleEngine.Run(ctx)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %v", err)
		}
	}()
	go grpcSrv.Serve(lis)

	connector.StartAll(cfg.Connectors, bus)

	hasConnector := cfg.Connectors.Modbus.Enabled || cfg.Connectors.OPCUA.Enabled
	if !hasConnector {
		logging.Logger.Println("no connectors enabled — starting demo producer")
		go demo(ctx, bus)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	logging.Logger.Println("shutting down…")

	// Give in-flight requests up to 10 s to complete
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
	}
	grpcSrv.GracefulStop()
	logging.Logger.Println("shutdown complete")
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
