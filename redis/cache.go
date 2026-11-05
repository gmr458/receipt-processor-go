package redis

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/receipt"
)

type Cache struct {
	Receipt receipt.ReceiptCache
	Session auth.SessionCache
}

func NewCache(redisClient *redis.Client) Cache {
	return Cache{
		Receipt: ReceiptCache{redisClient, 2 * time.Hour},
		Session: SessionCache{redisClient},
	}
}
