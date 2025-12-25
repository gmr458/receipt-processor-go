package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gmr458/receipt-processor/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
	DB                  *sql.DB
	ctx                 context.Context
	cancel              func()
	Dsn                 string
	logger              logger.Logger
	statsUpdateInterval time.Duration
}

func NewConn(dsn string, log logger.Logger, statsUpdateInterval time.Duration) (*Conn, error) {
	conn := &Conn{
		Dsn:                 dsn,
		logger:              log,
		statsUpdateInterval: statsUpdateInterval,
	}
	conn.ctx, conn.cancel = context.WithCancel(context.Background())

	if conn.Dsn == "" {
		return nil, fmt.Errorf("sqlite3 dsn required")
	}

	if conn.Dsn != ":memory:" {
		pathdir := filepath.Dir(conn.Dsn)
		if err := os.MkdirAll(pathdir, 0o700); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		conn.logger.Info("sqlite3 database file created")
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

	if _, err := conn.DB.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return nil, fmt.Errorf("error applying journal_mode = wal: %w", err)
	}
	conn.logger.Info("journal_mode = wal applied")

	if _, err := conn.DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}
	conn.logger.Info("foreign keys enabled")

	if _, err := conn.DB.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return nil, fmt.Errorf("error setting busy_timeout: %w", err)
	}
	conn.logger.Info("busy_timeout set to 5000ms")

	if err := conn.migrate(); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}
	conn.logger.Info("successful migration")

	if err := conn.DB.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite ping error: %w", err)
	}
	conn.logger.Info("successful ping")

	go conn.monitor()

	return conn, nil
}

func (conn *Conn) migrate() error {
	const query = `
		CREATE TABLE IF NOT EXISTS migrations (
			name TEXT PRIMARY KEY
		);
	`
	if _, err := conn.DB.Exec(query); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	names, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to glob migration files: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := conn.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}

	return nil
}

func (conn *Conn) migrateFile(name string) error {
	tx, err := conn.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var n int
	const queryCount = `
		SELECT COUNT(*)
		FROM migrations
		WHERE name = ?
	`
	if err := tx.QueryRow(queryCount, name).Scan(&n); err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	} else if n != 0 {
		return nil // migration already applied, skip
	}

	buf, err := fs.ReadFile(migrationsFS, name)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := tx.Exec(string(buf)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	const queryInsert = `INSERT INTO migrations (name) VALUES (?)`
	if _, err := tx.Exec(queryInsert, name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

func (conn *Conn) Close() error {
	conn.cancel()

	if conn.DB != nil {
		return conn.DB.Close()
	}

	return nil
}

func (conn *Conn) updateStats(ctx context.Context) error {
	tx, err := conn.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var n int

	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM receipt`).Scan(&n)
	if err != nil {
		return fmt.Errorf("failed to count receipts: %w", err)
	}
	receiptCountGauge.Set(float64(n))

	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM item`).Scan(&n)
	if err != nil {
		return fmt.Errorf("failed to count items: %w", err)
	}
	itemCountGauge.Set(float64(n))

	return nil
}

func (conn *Conn) monitor() {
	ticker := time.NewTicker(conn.statsUpdateInterval)
	defer ticker.Stop()

	conn.logger.Info("stats monitor started", "interval", conn.statsUpdateInterval)

	for {
		select {
		case <-conn.ctx.Done():
			conn.logger.Info("stats monitor stopped")
			return
		case <-ticker.C:
			if err := conn.updateStats(conn.ctx); err != nil {
				conn.logger.Error("error updating stats", "error", err)
			}
		}
	}
}
