package service

import (
	goredis "github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/redis"
	"github.com/gmr458/receipt-processor/sqlite"
)

type Service struct {
	Receipt ReceiptService
}

func New(conn *sqlite.Conn, redisClient *goredis.Client) Service {
	repository := sqlite.NewRepository(conn)
	cache := redis.NewCache(redisClient)

	return Service{
		Receipt: NewReceiptService(
			repository.Receipt,
			cache.Receipt,
		),
	}
}
