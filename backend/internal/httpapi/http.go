package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type server struct {
	db *sql.DB
}

func StartWithAddrCtx(ctx context.Context, db *sql.DB, addr string, apiKey string) {
	s := &server{db: db}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/items", s.items)

	handler := APIKey(apiKey)(RateLimit(10)(CORS(RequestID(Logging(mux)))))

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go srv.ListenAndServe()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

func (s *server) items(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, _ := s.db.Query("select id, name from items")
		var items []Item
		for rows.Next() {
			var i Item
			rows.Scan(&i.ID, &i.Name)
			items = append(items, i)
		}
		json.NewEncoder(w).Encode(items)

	case http.MethodPost:
		var i Item
		json.NewDecoder(r.Body).Decode(&i)
		s.db.Exec("insert into items(name) values(?)", i.Name)
		w.WriteHeader(http.StatusCreated)
	}
}
