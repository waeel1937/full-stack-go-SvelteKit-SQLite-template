package sync

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Aggregate struct {
	Time   int64   `json:"time"`
	Window int64   `json:"window_ms"`
	Metric string  `json:"metric"`
	Avg    float64 `json:"avg"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int     `json:"count"`
}

type CloudSync struct {
	DB        *sql.DB
	Endpoint  string
	Interval  time.Duration
}

func (c *CloudSync) Run() {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for range ticker.C {
		rows, err := c.DB.Query(`
SELECT time, window, metric, avg, min, max, count
FROM aggregates
ORDER BY time DESC
LIMIT 50
`)
		if err != nil {
			continue
		}

		var batch []Aggregate
		for rows.Next() {
			var a Aggregate
			if rows.Scan(
				&a.Time,
				&a.Window,
				&a.Metric,
				&a.Avg,
				&a.Min,
				&a.Max,
				&a.Count,
			) == nil {
				batch = append(batch, a)
			}
		}
		rows.Close()

		if len(batch) == 0 {
			continue
		}

		b, _ := json.Marshal(batch)
		http.Post(c.Endpoint, "application/json", bytes.NewReader(b))
	}
}
