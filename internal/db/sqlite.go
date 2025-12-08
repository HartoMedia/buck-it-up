package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping sqlite db: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`
        CREATE TABLE IF NOT EXISTS buckets (
          id          INTEGER PRIMARY KEY AUTOINCREMENT,
          name        TEXT NOT NULL UNIQUE,
          created_at  INTEGER NOT NULL
        );
        `,
		`
        CREATE TABLE IF NOT EXISTS access_keys (
          id           INTEGER PRIMARY KEY AUTOINCREMENT,
          bucket_id    INTEGER NOT NULL,
          key_id       TEXT NOT NULL,
          secret_hash  TEXT NOT NULL,
          role         TEXT NOT NULL,
          created_at   INTEGER NOT NULL,
          FOREIGN KEY(bucket_id) REFERENCES buckets(id),
          UNIQUE (key_id)
        );
        `,
		`
        CREATE TABLE IF NOT EXISTS objects (
          id            INTEGER PRIMARY KEY AUTOINCREMENT,
          bucket_id     INTEGER NOT NULL,
          object_key    TEXT NOT NULL,
          file_path     TEXT NOT NULL,
          size          INTEGER NOT NULL,
          content_type  TEXT,
          checksum      TEXT,
          created_at    INTEGER NOT NULL,
          FOREIGN KEY(bucket_id) REFERENCES buckets(id),
          UNIQUE(bucket_id, object_key)
        );
        `,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
