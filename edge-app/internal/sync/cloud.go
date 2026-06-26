package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"edge-app/internal/storage"
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
	Store    *storage.Store
	Endpoint string
	Interval time.Duration
}

func (c *CloudSync) Run(ctx context.Context) {
	lastSync := c.loadLastSync()

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	log.Printf("cloud sync started interval=%s lastSync=%d", c.Interval, lastSync)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastSync = c.sync(ctx, lastSync)
		}
	}
}

func (c *CloudSync) loadLastSync() int64 {
	v, ok, err := c.Store.KVGet("cloud_sync_last")
	if err != nil {
		log.Printf("cloud sync: could not load cursor: %v", err)
		return 0
	}
	if !ok {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func (c *CloudSync) sync(ctx context.Context, lastSync int64) int64 {
	rows, err := c.Store.DB.QueryContext(ctx,
		`SELECT time, window, metric, avg, min, max, count
		 FROM aggregates WHERE time > ? ORDER BY time ASC LIMIT 50`,
		lastSync,
	)
	if err != nil {
		log.Println("sync query error:", err)
		return lastSync
	}

	var batch []Aggregate
	var maxTime int64
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(&a.Time, &a.Window, &a.Metric, &a.Avg, &a.Min, &a.Max, &a.Count); err == nil {
			batch = append(batch, a)
			if a.Time > maxTime {
				maxTime = a.Time
			}
		}
	}
	rows.Close()

	if len(batch) == 0 {
		return lastSync
	}

	b, err := json.Marshal(batch)
	if err != nil {
		log.Printf("sync marshal error: %v", err)
		return lastSync
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(b))
	if err != nil {
		log.Printf("sync request error: %v", err)
		return lastSync
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("sync error: %v (will retry)", err)
		return lastSync
	}
	resp.Body.Close()

	if resp.StatusCode < 300 {
		if err := c.Store.KVSet("cloud_sync_last", strconv.FormatInt(maxTime, 10)); err != nil {
			log.Printf("sync: could not persist cursor: %v", err)
		}
		log.Printf("synced %d aggregates up to t=%d", len(batch), maxTime)
		return maxTime
	}

	log.Printf("sync: remote returned HTTP %d", resp.StatusCode)
	return lastSync
}
