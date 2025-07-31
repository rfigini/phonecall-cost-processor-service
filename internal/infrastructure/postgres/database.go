package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgresConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createCallsTableIfNotExists(db); err != nil {
		return nil, fmt.Errorf("error creating calls table: %w", err)
	}

	return db, nil
}

func createCallsTableIfNotExists(db *sql.DB) error {
	const query = `
	CREATE TABLE IF NOT EXISTS calls (
		call_id UUID PRIMARY KEY,
		caller TEXT,
		receiver TEXT,
		duration_in_seconds INT,
		start_timestamp TIMESTAMPTZ,
		cost NUMERIC(10, 2),
		currency TEXT,
		refunded BOOLEAN DEFAULT false,
		refund_reason TEXT,
		status VARCHAR(20) DEFAULT 'PENDING' NOT NULL,
		processed_at TIMESTAMPTZ DEFAULT now() NOT NULL,
		CONSTRAINT unique_call_id UNIQUE (call_id)
	);`
	_, err := db.Exec(query)
	return err
}
