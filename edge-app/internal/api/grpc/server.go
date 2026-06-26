package grpcapi

import (
	"context"
	"database/sql"

	"edge-app/internal/core"
	pb "edge-app/internal/api/grpc/pb"
)

type Server struct {
	pb.UnimplementedEdgeServiceServer
	DB  *sql.DB
	Bus *core.Bus
}

// StreamMetrics subscribes to the fan-out event bus and pushes every
// MetricEvent to the gRPC client as a server-side stream. The stream
// closes when the client disconnects or the server shuts down.
func (s *Server) StreamMetrics(req *pb.StreamRequest, stream pb.EdgeService_StreamMetricsServer) error {
	ch := s.Bus.SubscribeMetrics(1024)
	filter := req.GetKeyFilter()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case m, ok := <-ch:
			if !ok {
				return nil
			}
			if filter != "" && m.Key != filter {
				continue
			}
			if err := stream.Send(&pb.MetricEvent{
				Time:   m.Time.UnixMilli(),
				Source: m.Source,
				Key:    m.Key,
				Value:  m.Value,
				Ok:     m.OK,
			}); err != nil {
				return err
			}
		}
	}
}

// GetAggregates returns the most recent aggregates for a given window size.
func (s *Server) GetAggregates(ctx context.Context, req *pb.AggregateRequest) (*pb.AggregateResponse, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT time, window, metric, avg, min, max, count
		FROM aggregates
		WHERE window = ?
		ORDER BY time DESC
		LIMIT 100
	`, req.WindowMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := &pb.AggregateResponse{}
	for rows.Next() {
		a := &pb.Aggregate{}
		if err := rows.Scan(
			&a.Time, &a.WindowMs, &a.Metric,
			&a.Avg, &a.Min, &a.Max, &a.Count,
		); err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, a)
	}
	return resp, rows.Err()
}
