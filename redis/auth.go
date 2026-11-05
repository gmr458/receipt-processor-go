package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/errs"
)

const (
	sessionRecordKeyPrefix = "auth:session:token:"
	sessionIndexKeyPrefix  = "auth:session:tokens:"
)

type SessionCache struct {
	redisClient *redis.Client
}

func (c SessionCache) Get(ctx context.Context, token string) (*auth.Session, error) {
	val, err := c.redisClient.Get(ctx, sessionRecordKeyPrefix+token).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, &errs.Error{
				Code:    errs.ENOTFOUND,
				Message: "Session not found in cache",
			}
		default:
			return nil, err
		}
	}

	var session auth.Session
	err = json.Unmarshal([]byte(val), &session)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.EINTERNAL,
			Message: "Error unmarshaling session from redis",
		}
	}

	return &session, nil
}

func (c SessionCache) Set(
	ctx context.Context,
	session *auth.Session,
	ttl time.Duration,
) error {
	b, err := json.Marshal(session)
	if err != nil {
		return &errs.Error{
			Code:    errs.EINTERNAL,
			Message: "Error marshaling session before storing on redis",
		}
	}

	pipe := c.redisClient.TxPipeline()
	pipe.Set(ctx, sessionRecordKeyPrefix+session.Token, b, ttl)
	pipe.SAdd(ctx, sessionIndexKeyPrefix+session.UserID, session.Token)
	pipe.Expire(ctx, sessionIndexKeyPrefix+session.UserID, ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c SessionCache) Delete(ctx context.Context, token string) error {
	return c.redisClient.Del(ctx, sessionRecordKeyPrefix+token).Err()
}

func (c SessionCache) DeleteByUserID(ctx context.Context, userID string) error {
	key := sessionIndexKeyPrefix + userID
	tokens, err := c.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	args := make([]string, 0, len(tokens)+1)
	args = append(args, key)
	for _, token := range tokens {
		args = append(args, sessionRecordKeyPrefix+token)
	}

	return c.redisClient.Del(ctx, args...).Err()
}
