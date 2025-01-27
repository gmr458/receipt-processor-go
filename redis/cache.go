package redis

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/domain"
)

type Cache struct {
	Receipt domain.ReceiptCache
}

func NewCache(redisClient *redis.Client) Cache {
	return Cache{
		Receipt: ReceiptCache{redisClient, 2 * time.Hour},
	}
}
