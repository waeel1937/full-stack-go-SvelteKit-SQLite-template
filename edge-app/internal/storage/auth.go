package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserExists = errors.New("user already exists")

func (s *Store) InitAuth() error {
	_, err := s.DB.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	salt TEXT NOT NULL,
	created_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	username TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);`)
	return err
}

func hashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + password))
	return hex.EncodeToString(h.Sum(nil))
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) CreateUser(username, password string) error {
	var count int
	s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username=?", username).Scan(&count)
	if count > 0 {
		return ErrUserExists
	}
	salt := randomHex(16)
	hash := hashPassword(password, salt)
	_, err := s.DB.Exec(
		"INSERT INTO users(username, password_hash, salt, created_at) VALUES(?,?,?,?)",
		username, hash, salt, time.Now().Unix(),
	)
	return err
}

func (s *Store) Authenticate(username, password string) (string, error) {
	var hash, salt string
	err := s.DB.QueryRow("SELECT password_hash, salt FROM users WHERE username=?", username).Scan(&hash, &salt)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if hashPassword(password, salt) != hash {
		return "", ErrInvalidCredentials
	}
	token := randomHex(32)
	expires := time.Now().Add(24 * time.Hour).Unix()
	s.DB.Exec(
		"INSERT INTO sessions(token, username, created_at, expires_at) VALUES(?,?,?,?)",
		token, username, time.Now().Unix(), expires,
	)
	return token, nil
}

func (s *Store) ValidateSession(token string) (string, bool) {
	var username string
	var expires int64
	err := s.DB.QueryRow("SELECT username, expires_at FROM sessions WHERE token=?", token).Scan(&username, &expires)
	if err != nil {
		return "", false
	}
	if time.Now().Unix() > expires {
		s.DB.Exec("DELETE FROM sessions WHERE token=?", token)
		return "", false
	}
	return username, true
}

func (s *Store) DeleteSession(token string) {
	s.DB.Exec("DELETE FROM sessions WHERE token=?", token)
}
