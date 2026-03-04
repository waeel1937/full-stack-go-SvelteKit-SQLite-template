package api

import (
	"encoding/json"
	"net/http"

	"edge-app/internal/storage/ringbuffer"
)

type RawServer struct {
	Buffer *ringbuffer.Buffer
}

func (r *RawServer) Handler(w http.ResponseWriter, _ *http.Request) {
	data := r.Buffer.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
