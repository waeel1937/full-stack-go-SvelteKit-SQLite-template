package grpcapi

import (
	"context"
	"log"
	"net"
	"time"

	"backend/proto"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedHealthServer
}

func (s *server) Check(ctx context.Context, _ *proto.Empty) (*proto.Status, error) {
	return &proto.Status{Value: "ok"}, nil
}

func StartCtx(ctx context.Context, lis net.Listener) {
	g := grpc.NewServer()
	proto.RegisterHealthServer(g, &server{})

	go func() {
		<-ctx.Done()
		stopped := make(chan struct{})
		go func() {
			g.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			g.Stop()
		}
	}()

	log.Fatal(g.Serve(lis))
}
