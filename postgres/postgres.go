package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Conn struct {
	Pool   *pgxpool.Pool
	Dsn    string
	logger *slog.Logger
}

func NewConn(dsn string, logger *slog.Logger) (*Conn, error) {
	conn := &Conn{
		Dsn:    dsn,
		logger: logger,
	}

	if conn.Dsn == "" {
		return nil, fmt.Errorf("postgres dsn required")
	}
	conn.logger.Info("using postgres database")

	config, err := pgxpool.ParseConfig(conn.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}
	conn.Pool = pool

	if err := conn.Pool.Ping(context.Background()); err != nil {
		conn.Pool.Close()
		return nil, fmt.Errorf("postgres ping error: %w", err)
	}
	conn.logger.Info("postgres connection established")

	return conn, nil
}

func (conn *Conn) Close() error {
	if conn.Pool != nil {
		conn.Pool.Close()
	}

	return nil
}
