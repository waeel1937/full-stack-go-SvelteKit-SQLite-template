package storage

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type RuleRecord struct {
	ID        string
	Key       string
	Condition string
	Threshold float64
	Message   string
	Enabled   bool
}

type Store struct {
	DB     *sql.DB
	insAgg *sql.Stmt
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	pragmas := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA synchronous=NORMAL;`,
		`PRAGMA cache_size=-32000;`,
		`PRAGMA foreign_keys=ON;`,
		`PRAGMA busy_timeout=5000;`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, err
		}
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS aggregates (
	time    INTEGER NOT NULL,
	window  INTEGER NOT NULL,
	metric  TEXT    NOT NULL,
	avg     REAL    NOT NULL,
	min     REAL    NOT NULL,
	max     REAL    NOT NULL,
	count   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agg_window_time ON aggregates(window, time DESC);

CREATE TABLE IF NOT EXISTS rules (
	id        TEXT PRIMARY KEY,
	key       TEXT    NOT NULL,
	condition TEXT    NOT NULL,
	threshold REAL    NOT NULL,
	message   TEXT    NOT NULL,
	enabled   INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS kv (
	k TEXT PRIMARY KEY,
	v TEXT NOT NULL
);
`); err != nil {
		return nil, err
	}

	insAgg, err := db.Prepare(
		`INSERT INTO aggregates(time, window, metric, avg, min, max, count) VALUES(?,?,?,?,?,?,?)`,
	)
	if err != nil {
		return nil, err
	}

	return &Store{DB: db, insAgg: insAgg}, nil
}

func (s *Store) Close() error {
	s.insAgg.Close()
	return s.DB.Close()
}

func (s *Store) InsertAggregate(t time.Time, window time.Duration, metric string, avg, min, max float64, count int) error {
	_, err := s.insAgg.Exec(t.Unix(), int64(window/time.Millisecond), metric, avg, min, max, count)
	return err
}

func (s *Store) LoadRules() ([]RuleRecord, error) {
	rows, err := s.DB.Query(`SELECT id, key, condition, threshold, message, enabled FROM rules`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RuleRecord
	for rows.Next() {
		var r RuleRecord
		var enabled int
		if err := rows.Scan(&r.ID, &r.Key, &r.Condition, &r.Threshold, &r.Message, &enabled); err != nil {
			return nil, err
		}
		r.Enabled = enabled != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) UpsertRule(id, key, condition string, threshold float64, message string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := s.DB.Exec(
		`INSERT INTO rules(id, key, condition, threshold, message, enabled) VALUES(?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET
		   key=excluded.key, condition=excluded.condition,
		   threshold=excluded.threshold, message=excluded.message,
		   enabled=excluded.enabled`,
		id, key, condition, threshold, message, enabledInt,
	)
	return err
}

func (s *Store) KVGet(key string) (string, bool, error) {
	var val string
	err := s.DB.QueryRow(`SELECT v FROM kv WHERE k = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (s *Store) KVSet(key, val string) error {
	_, err := s.DB.Exec(
		`INSERT INTO kv(k,v) VALUES(?,?) ON CONFLICT(k) DO UPDATE SET v=excluded.v`,
		key, val,
	)
	return err
}
