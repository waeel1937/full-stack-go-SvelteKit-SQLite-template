package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS aggregates (
	time INTEGER NOT NULL,
	window INTEGER NOT NULL,
	metric TEXT NOT NULL,
	avg REAL NOT NULL,
	min REAL NOT NULL,
	max REAL NOT NULL,
	count INTEGER NOT NULL
);
`); err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) InsertAggregate(
	t time.Time,
	window time.Duration,
	metric string,
	avg, min, max float64,
	count int,
) error {
	_, err := s.DB.Exec(
		`INSERT INTO aggregates(time, window, metric, avg, min, max, count)
		 VALUES(?,?,?,?,?,?,?)`,
		t.Unix(),
		int64(window/time.Millisecond),
		metric,
		avg,
		min,
		max,
		count,
	)
	return err
}
