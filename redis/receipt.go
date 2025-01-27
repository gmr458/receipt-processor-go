package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/domain"
)

type ReceiptCache struct {
	redisClient *redis.Client
	duration    time.Duration
}

func (c ReceiptCache) GetPointsById(ctx context.Context, id string) (int, error) {
	points, err := c.redisClient.Get(ctx, id).Int()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return 0, &domain.Error{
				Code:    domain.ENOTFOUND,
				Message: "Receipt's points not found in cache",
			}
		default:
			return 0, err
		}
	}

	return points, nil
}

func (c ReceiptCache) SetPointsById(ctx context.Context, id string, points int) error {
	err := c.redisClient.Set(ctx, id, points, c.duration).Err()
	if err != nil {
		return err
	}

	return nil
}
