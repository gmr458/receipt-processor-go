package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Conn struct {
	DB     *sql.DB
	Dsn    string
	logger *slog.Logger
}

func NewConn(dsn string, log *slog.Logger) (*Conn, error) {
	conn := &Conn{
		Dsn:    dsn,
		logger: log,
	}

	if conn.Dsn == "" {
		return nil, fmt.Errorf("sqlite3 dsn required")
	}

	if conn.Dsn != ":memory:" {
		pathdir := filepath.Dir(conn.Dsn)
		if err := os.MkdirAll(pathdir, 0o700); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		conn.logger.Info("sqlite3 database directory ready")
	} else {
		conn.logger.Info("using sqlite3 in memory database")
	}

	var err error
	conn.DB, err = sql.Open("sqlite3", conn.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite3 database: %w", err)
	}
	conn.logger.Info("sqlite3 database connection opened")

	conn.DB.SetMaxOpenConns(1)
	conn.DB.SetMaxIdleConns(1)
	conn.DB.SetConnMaxLifetime(0)

	if _, err := conn.DB.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return nil, fmt.Errorf("error setting busy_timeout: %w", err)
	}
	conn.logger.Info("busy_timeout set to 5000ms")

	if _, err := conn.DB.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return nil, fmt.Errorf("error applying journal_mode = wal: %w", err)
	}
	conn.logger.Info("journal_mode = wal applied")

	if _, err := conn.DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}
	conn.logger.Info("foreign keys enabled")

	if err := conn.DB.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite ping error: %w", err)
	}
	conn.logger.Info("successful ping")

	return conn, nil
}

func (conn *Conn) Close() error {
	if conn.DB != nil {
		return conn.DB.Close()
	}

	return nil
}
