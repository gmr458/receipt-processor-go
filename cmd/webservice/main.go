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

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/env"
	"github.com/gmr458/receipt-processor/mailer"
	"github.com/gmr458/receipt-processor/postgres"
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

	cfg.db.dsn = env.GetenvOrDefault("DSN", "postgres://postgres:postgres@localhost:5432/receipt_processor?sslmode=disable")

	trustedOrigins := env.Getenv[string]("CORS_TRUSTED_ORIGINS")
	cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg.limiter.enabled = env.GetenvOrDefault("LIMITER_ENABLED", true)
	cfg.limiter.rps = env.GetenvOrDefault("LIMITER_RPS", 10.0)
	cfg.limiter.burst = env.GetenvOrDefault("LIMITER_BURST", 20)
	cfg.limiter.trustedProxyHeader = env.GetenvOrDefault("TRUSTED_PROXY_HEADER", "")

	cfg.auth.cookieName = env.GetenvOrDefault("AUTH_SESSION_COOKIE_NAME", "better-auth.session_token")
	cfg.auth.sessionDuration = time.Duration(env.GetenvOrDefault("AUTH_SESSION_DURATION_HOURS", 168)) * time.Hour
	cfg.auth.rememberMeDuration = time.Duration(env.GetenvOrDefault("AUTH_REMEMBER_ME_DURATION_HOURS", 720)) * time.Hour
	cfg.auth.verificationExpiry = time.Duration(env.GetenvOrDefault("AUTH_VERIFICATION_EXPIRY_HOURS", 1)) * time.Hour
	cfg.auth.defaultCallbackURL = env.GetenvOrDefault("AUTH_CALLBACK_URL", "")

	cfg.smtp.host = env.GetenvOrDefault("SMTP_HOST", "localhost")
	cfg.smtp.port = env.GetenvOrDefault("SMTP_PORT", 1025)
	cfg.smtp.username = env.GetenvOrDefault("SMTP_USERNAME", "")
	cfg.smtp.password = env.GetenvOrDefault("SMTP_PASSWORD", "")
	cfg.smtp.fromAddr = env.GetenvOrDefault("MAIL_FROM_ADDR", "noreply@receipt-processor.local")
	cfg.smtp.fromName = env.GetenvOrDefault("MAIL_FROM_NAME", "Receipt Processor")

	cfg.redis.addr = env.GetenvOrDefault("REDIS_ADDR", "localhost:6379")
	cfg.redis.password = env.Getenv[string]("REDIS_PASSWORD")
	cfg.redis.db = env.GetenvOrDefault("REDIS_DB", 0)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	postgresConn, err := postgres.NewConn(cfg.db.dsn, logger, 15*time.Second)
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

	app := newApp(
		cfg,
		logger,
		postgresConn,
		redisClient,
		mustNewMailer(cfg, logger),
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

	if err := postgresConn.Close(); err != nil {
		logger.Error("failed to close postgres connection", "error", err)
	}
	if err := redisClient.Close(); err != nil {
		logger.Error("failed to close redis connection", "error", err)
	}
}

func mustNewMailer(cfg config, logger *slog.Logger) auth.Mailer {
	m, err := mailer.New(mailer.Config{
		Host:     cfg.smtp.host,
		Port:     cfg.smtp.port,
		Username: cfg.smtp.username,
		Password: cfg.smtp.password,
		FromAddr: cfg.smtp.fromAddr,
		FromName: cfg.smtp.fromName,
	})
	if err != nil {
		logger.Error("failed to create mailer", "error", err)
		os.Exit(1)
	}

	logger.Info("mailer created", "host", cfg.smtp.host, "port", cfg.smtp.port)

	return m
}
