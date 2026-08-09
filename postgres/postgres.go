package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	receiptCountGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "receipt_processor",
			Name:      "receipt_count",
			Help:      "Number of receipts created.",
		},
	)

	itemCountGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "receipt_processor",
			Name:      "item_count",
			Help:      "Number of items created.",
		},
	)
)

type Conn struct {
	Pool                *pgxpool.Pool
	ctx                 context.Context
	cancel              func()
	Dsn                 string
	logger              *slog.Logger
	statsUpdateInterval time.Duration
}

func NewConn(dsn string, logger *slog.Logger, statsUpdateInterval time.Duration) (*Conn, error) {
	conn := &Conn{
		Dsn:                 dsn,
		logger:              logger,
		statsUpdateInterval: statsUpdateInterval,
	}
	conn.ctx, conn.cancel = context.WithCancel(context.Background())

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

	go conn.monitor()

	return conn, nil
}

func (conn *Conn) Close() error {
	conn.cancel()

	if conn.Pool != nil {
		conn.Pool.Close()
	}

	return nil
}

func (conn *Conn) updateStats(ctx context.Context) error {
	tx, err := conn.Pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var n int

	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM receipt`).Scan(&n)
	if err != nil {
		return fmt.Errorf("failed to count receipts: %w", err)
	}
	receiptCountGauge.Set(float64(n))

	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM item`).Scan(&n)
	if err != nil {
		return fmt.Errorf("failed to count items: %w", err)
	}
	itemCountGauge.Set(float64(n))

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (conn *Conn) monitor() {
	ticker := time.NewTicker(conn.statsUpdateInterval)
	defer ticker.Stop()

	conn.logger.Info("postgres stats monitor started", "interval", conn.statsUpdateInterval)

	for {
		select {
		case <-conn.ctx.Done():
			conn.logger.Info("postgres stats monitor stopped")
			return
		case <-ticker.C:
			if err := conn.updateStats(conn.ctx); err != nil {
				conn.logger.Error("error updating postgres stats", "error", err)
			}
		}
	}
}
