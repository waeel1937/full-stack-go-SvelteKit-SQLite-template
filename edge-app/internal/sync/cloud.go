package sync

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
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
	DB       *sql.DB
	Endpoint string
	Interval time.Duration
	lastSync int64
}

func (c *CloudSync) Run() {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	log.Printf("cloud sync started interval=%s", c.Interval)

	for range ticker.C {
		rows, err := c.DB.Query(
			"SELECT time, window, metric, avg, min, max, count FROM aggregates WHERE time > ? ORDER BY time ASC LIMIT 50",
			c.lastSync,
		)
		if err != nil {
			log.Println("sync query error:", err)
			continue
		}

		var batch []Aggregate
		var maxTime int64
		for rows.Next() {
			var a Aggregate
			if rows.Scan(&a.Time, &a.Window, &a.Metric, &a.Avg, &a.Min, &a.Max, &a.Count) == nil {
				batch = append(batch, a)
				if a.Time > maxTime {
					maxTime = a.Time
				}
			}
		}
		rows.Close()

		if len(batch) == 0 {
			continue
		}

		b, _ := json.Marshal(batch)
		resp, err := http.Post(c.Endpoint, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Printf("sync error: %v (will retry)", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 300 {
			c.lastSync = maxTime
			log.Printf("synced %d aggregates up to t=%d", len(batch), maxTime)
		}
	}
}
