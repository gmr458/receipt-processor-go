package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/env"
	"github.com/gmr458/receipt-processor/sqlite"
)

var version string

func main() {
	displayVersion := flag.Bool("version", false, "Display version")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("version: %s\n", version)
		os.Exit(0)
	}

	var cfg config

	cfg.host = env.GetenvOrDefault("HOST", "127.0.0.1")
	cfg.port = env.GetenvOrDefault("PORT", 4000)

	cfg.debugPort = env.GetenvOrDefault("DEBUG_PORT", 4001)

	cfg.env = env.GetenvOrDefault("ENV", "development")

	cfg.db.dsn = env.GetenvOrDefault("DSN", ":memory:")

	trustedOrigins := env.Getenv[string]("CORS_TRUSTED_ORIGINS")
	cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg.limiter.enabled = env.GetenvOrDefault("LIMITER_ENABLED", true)
	cfg.limiter.rps = env.GetenvOrDefault("LIMITER_RPS", 10.0)
	cfg.limiter.burst = env.GetenvOrDefault("LIMITER_BURST", 20)
	cfg.limiter.trustedProxyHeader = env.GetenvOrDefault("TRUSTED_PROXY_HEADER", "")

	cfg.redis.addr = env.GetenvOrDefault("REDIS_ADDR", "localhost:6379")
	cfg.redis.password = env.Getenv[string]("REDIS_PASSWORD")
	cfg.redis.db = env.GetenvOrDefault("REDIS_DB", 0)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	sqliteConn, err := sqlite.NewConn(cfg.db.dsn, logger, 15*time.Second)
	if err != nil {
		logger.Error("failed to create sqlite connection", "error", err)
		os.Exit(1)
	}
	logger.Info("sqlite3 connection established")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.redis.addr,
		Password: cfg.redis.password,
		DB:       cfg.redis.db,
	})
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("failed to ping redis", "error", err)
		os.Exit(1)
	}
	logger.Info("redis connection established")

	app := newApp(
		cfg,
		logger,
		sqliteConn,
		redisClient,
	)

	go func() {
		err := app.serveDebug()
		if err != nil {
			logger.Error("debug server error", "error", err)
			os.Exit(1)
		}
	}()

	err = app.serve()
	if err != nil {
		logger.Error("main server error", "error", err)
		os.Exit(1)
	}

	if err := sqliteConn.Close(); err != nil {
		logger.Error("failed to close sqlite connection", "error", err)
	}
	if err := redisClient.Close(); err != nil {
		logger.Error("failed to close redis connection", "error", err)
	}
}
