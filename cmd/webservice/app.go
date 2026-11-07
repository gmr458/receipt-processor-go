package main

import (
	"log/slog"
	"net/http"
	"sync"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/cors"

	"github.com/gmr458/receipt-processor/postgres"
	"github.com/gmr458/receipt-processor/receipt"
	"github.com/gmr458/receipt-processor/redis"
)

type app struct {
	config         config
	logger         *slog.Logger
	server         *http.Server
	receiptService receipt.Service
	wg             sync.WaitGroup
	corsHandler    *cors.Cors
	rateLimiter    *redis.TokenBucket
}

func newApp(cfg config, logger *slog.Logger, postgresConn *postgres.Conn, redisClient *goredis.Client) *app {
	repository := postgres.NewRepository(postgresConn)
	cache := redis.NewCache(redisClient)

	return &app{
		config: cfg,
		logger: logger,
		receiptService: receipt.NewService(
			repository.Receipt,
			cache.Receipt,
		),
		corsHandler: cors.New(cors.Options{
			AllowedOrigins:   cfg.cors.trustedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           300,
		}),
		rateLimiter: redis.NewTokenBucket(redisClient, cfg.limiter.rps, cfg.limiter.burst),
	}
}
