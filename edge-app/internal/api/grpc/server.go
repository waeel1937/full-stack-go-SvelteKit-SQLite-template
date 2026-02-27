package grpcapi

import (
	"context"
	"database/sql"

	pb "edge-app/internal/api/grpc/pb"
)

type Server struct {
	pb.UnimplementedEdgeServiceServer
	DB *sql.DB
}

func (s *Server) GetAggregates(ctx context.Context, req *pb.AggregateRequest) (*pb.AggregateResponse, error) {
	rows, err := s.DB.Query(`
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
			&a.Time,
			&a.WindowMs,
			&a.Metric,
			&a.Avg,
			&a.Min,
			&a.Max,
			&a.Count,
		); err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, a)
	}
	return resp, nil
}
