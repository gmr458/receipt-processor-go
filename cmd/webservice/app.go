package main

import (
	"log/slog"
	"net/http"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/cors"

	"github.com/gmr458/receipt-processor-go/postgres"
	"github.com/gmr458/receipt-processor-go/receipt"
	"github.com/gmr458/receipt-processor-go/redis"
)

type app struct {
	config         config
	logger         *slog.Logger
	server         *http.Server
	receiptService receipt.Service
	corsHandler    *cors.Cors
	rateLimiter    *redis.TokenBucket
}

func newApp(cfg config, logger *slog.Logger, postgresConn *postgres.Conn, redisClient *goredis.Client) (*app, error) {
	rateLimiter, err := redis.NewTokenBucket(redisClient, cfg.limiter.rps, cfg.limiter.burst)
	if err != nil {
		return nil, err
	}

	repository := postgres.NewRepository(postgresConn)

	return &app{
		config:         cfg,
		logger:         logger,
		receiptService: receipt.NewService(repository.Receipt),
		corsHandler: cors.New(cors.Options{
			AllowedOrigins:   cfg.cors.trustedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           300,
		}),
		rateLimiter: rateLimiter,
	}, nil
}
