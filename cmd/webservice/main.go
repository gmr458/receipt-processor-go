package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor-go/env"
	"github.com/gmr458/receipt-processor-go/postgres"
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
	var err error

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	fail := func(err error) {
		logger.Error("invalid environment configuration", "error", err)
		os.Exit(1)
	}

	cfg.host, err = env.GetenvOrDefault("HOST", "127.0.0.1")
	if err != nil {
		fail(err)
	}
	cfg.port, err = env.GetenvOrDefault("PORT", 4000)
	if err != nil {
		fail(err)
	}

	cfg.env, err = env.GetenvOrDefault("ENV", "development")
	if err != nil {
		fail(err)
	}

	cfg.db.dsn, err = env.GetenvOrDefault("DSN", "postgres://postgres:postgres@localhost:5432/receipt_processor?sslmode=disable")
	if err != nil {
		fail(err)
	}

	trustedOrigins, err := env.Getenv[string]("CORS_TRUSTED_ORIGINS")
	if err != nil {
		fail(err)
	}
	cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg.limiter.enabled, err = env.GetenvOrDefault("LIMITER_ENABLED", true)
	if err != nil {
		fail(err)
	}
	cfg.limiter.rps, err = env.GetenvOrDefault("LIMITER_RPS", 10.0)
	if err != nil {
		fail(err)
	}
	cfg.limiter.burst, err = env.GetenvOrDefault("LIMITER_BURST", 20)
	if err != nil {
		fail(err)
	}
	cfg.limiter.trustedProxyHeader, err = env.GetenvOrDefault("TRUSTED_PROXY_HEADER", "")
	if err != nil {
		fail(err)
	}

	cfg.redis.addr, err = env.GetenvOrDefault("REDIS_ADDR", "localhost:6379")
	if err != nil {
		fail(err)
	}
	cfg.redis.password, err = env.GetenvOrDefault("REDIS_PASSWORD", "")
	if err != nil {
		fail(err)
	}
	cfg.redis.db, err = env.GetenvOrDefault("REDIS_DB", 0)
	if err != nil {
		fail(err)
	}

	postgresConn, err := postgres.NewConn(cfg.db.dsn, logger)
	if err != nil {
		logger.Error("failed to create postgres connection", "error", err)
		os.Exit(1)
	}
	logger.Info("postgres connection established")

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

	app, err := newApp(
		cfg,
		logger,
		postgresConn,
		redisClient,
	)
	if err != nil {
		logger.Error("failed to create application", "error", err)
		os.Exit(1)
	}

	err = app.serve()
	if err != nil {
		logger.Error("main server error", "error", err)
		os.Exit(1)
	}

	if err := postgresConn.Close(); err != nil {
		logger.Error("failed to close postgres connection", "error", err)
	}
	if err := redisClient.Close(); err != nil {
		logger.Error("failed to close redis connection", "error", err)
	}
}
