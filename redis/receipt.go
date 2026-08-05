package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/errs"
	"github.com/gmr458/receipt-processor/receipt"
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
			return 0, &errs.Error{
				Code:    errs.ENOTFOUND,
				Message: "Receipt's points not found in cache",
			}
		default:
			return 0, err
		}
	}

	return points, nil
}

func (c ReceiptCache) SetPointsById(
	ctx context.Context,
	id string,
	points int,
	exp time.Duration,
) error {
	return c.redisClient.Set(
		ctx,
		id,
		points,
		exp,
	).Err()
}

func (c ReceiptCache) SetPaginatedReceipts(
	ctx context.Context,
	key string,
	paginatedReceipts receipt.PaginatedReceipts,
	exp time.Duration,
) error {
	b, err := json.Marshal(paginatedReceipts)
	if err != nil {
		return &errs.Error{
			Code:    errs.EINTERNAL,
			Message: "Error marshaling paginated receipts before storing on redis",
		}
	}
	err = c.redisClient.Set(ctx, key, b, exp).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c ReceiptCache) GetPaginatedReceipts(ctx context.Context, key string) (receipt.PaginatedReceipts, error) {
	val, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return receipt.PaginatedReceipts{}, &errs.Error{
				Code:    errs.ENOTFOUND,
				Message: "Paginated receipts not found in cache",
			}
		default:
			return receipt.PaginatedReceipts{}, err
		}
	}

	var result receipt.PaginatedReceipts
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		return receipt.PaginatedReceipts{}, &errs.Error{
			Code:    errs.EINTERNAL,
			Message: "Error unmarshaling paginated receipts from redis",
		}
	}

	return result, nil
}
