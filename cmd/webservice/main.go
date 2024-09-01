package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/cache"
	"github.com/gmr458/receipt-processor/env"
	"github.com/gmr458/receipt-processor/service"
	"github.com/gmr458/receipt-processor/sqlite"
)

func main() {
	var cfg config

	cfg.port = env.GetenvInt("PORT")
	cfg.debugPort = env.GetenvInt("DEBUG_PORT")

	cfg.env = env.Getenv("ENV")

	cfg.db.dsn = env.Getenv("DSN")

	trustedOrigins := env.Getenv("CORS_TRUSTED_ORIGINS")
	cfg.cors.trustedOrigins = strings.Fields(trustedOrigins)

	cfg.limiter.enabled = env.GetenvBool("LIMITER_ENABLED")
	cfg.limiter.rps = env.GetenvFloat("LIMITER_RPS")
	cfg.limiter.burst = env.GetenvInt("LIMITER_BURST")

	cfg.redis.addr = env.Getenv("REDIS_ADDR")
	cfg.redis.password = env.Getenv("REDIS_PASSWORD")
	cfg.redis.db = env.GetenvInt("REDIS_DB")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	sqliteConn, err := sqlite.NewConn(cfg.db.dsn, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("sqlite3 connection stablished")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.redis.addr,
		Password: cfg.redis.password,
		DB:       cfg.redis.db,
	})
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("redis connection stablished")

	app := &app{
		config:  cfg,
		logger:  logger,
		service: service.New(sqliteConn),
		cache:   cache.New(redisClient),
	}

	go func() {
		err := app.serveDebug()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}()

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
