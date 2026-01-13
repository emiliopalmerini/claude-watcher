package database

import (
	"database/sql"
	"time"

	_ "github.com/tursodatabase/go-libsql"
)

func NewTurso(databaseURL, authToken string) (*sql.DB, error) {
	connStr := databaseURL + "?authToken=" + authToken
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent stale connections.
	// Turso's Hrana protocol closes idle streams, so we need to
	// ensure connections are refreshed before they become stale.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
