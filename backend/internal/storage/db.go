package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func OpenWithPath(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		create table if not exists items (
			id integer primary key autoincrement,
			name text not null
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
