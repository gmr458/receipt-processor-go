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

	_ "github.com/mattn/go-sqlite3"

	"github.com/gmr458/receipt-processor/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Conn struct {
	DB     *sql.DB
	ctx    context.Context
	cancel func()
	Dsn    string
}

func NewConn(dsn string, log logger.Logger) (*Conn, error) {
	conn := &Conn{Dsn: dsn}
	conn.ctx, conn.cancel = context.WithCancel(context.Background())

	if conn.Dsn == "" {
		err := fmt.Errorf("sqlite3 dsn required")
		log.Error(err.Error())
		return nil, err
	}

	if conn.Dsn != ":memory:" {
		pathdir := filepath.Dir(conn.Dsn)
		if err := os.MkdirAll(pathdir, 0700); err != nil {
			log.Error(err.Error())
			return nil, err
		}
		log.Info("sqlite3 database file created")
	} else {
		log.Info("using sqlite3 in memory database")
	}

	var err error
	conn.DB, err = sql.Open("sqlite3", conn.Dsn)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Info("sqlite3 database connection opened")

	if _, err := conn.DB.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		err = fmt.Errorf("error applying journal_mode = wal: %w", err)
		log.Error(err.Error())
		return nil, err
	}
	log.Info("journal_mode = wal applied")

	if _, err := conn.DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		err = fmt.Errorf("error enabling foreign keys: %w", err)
		log.Error(err.Error())
		return nil, err
	}
	log.Info("foreign keys enabled")

	if err := conn.migrate(); err != nil {
		err = fmt.Errorf("migration error: %w", err)
		log.Error(err.Error())
		return nil, err
	}
	log.Info("successful migration")

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
		return err
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
		return err
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
		return err
	} else if n != 0 { // already run migration, skip
		return nil
	}

	if buf, err := fs.ReadFile(migrationsFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	const queryInsert = `INSERT INTO migrations (name) VALUES (?)`
	if _, err := tx.Exec(queryInsert, name); err != nil {
		return err
	}

	return tx.Commit()
}

func (conn *Conn) Close() error {
	conn.cancel()

	if conn.DB != nil {
		return conn.DB.Close()
	}

	return nil
}
