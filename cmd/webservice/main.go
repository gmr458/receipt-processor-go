package main

import (
	"log/slog"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"

	"github.com/gmr458/receipt-processor/env"
	"github.com/gmr458/receipt-processor/service"
	"github.com/gmr458/receipt-processor/sqlite"
)

func main() {
	var cfg config

	cfg.port = env.GetenvInt("PORT")
	cfg.env = env.Getenv("ENV")

	cfg.db.dsn = env.Getenv("DSN")

	trustedOrigins := env.Getenv("CORS_TRUSTED_ORIGINS")
	cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg.limiter.enabled = env.GetenvBool("LIMITER_ENABLED")
	cfg.limiter.rps = env.GetenvFloat("LIMITER_RPS")
	cfg.limiter.burst = env.GetenvInt("LIMITER_BURST")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	sqliteConn, err := sqlite.NewConn(cfg.db.dsn, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("sqlite3 connection stablished")

	app := &app{
		config:  cfg,
		logger:  logger,
		service: service.New(sqliteConn),
	}
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
