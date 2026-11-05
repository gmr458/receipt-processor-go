package auth

import (
	"context"
	"time"
)

type Session struct {
	ID             string    `json:"id"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Token          string    `json:"token"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	IPAddress      *string   `json:"ipAddress"`
	UserAgent      *string   `json:"userAgent"`
	UserID         string    `json:"userId"`
	ImpersonatedBy *string   `json:"impersonatedBy"`
}

func (s *Session) IsExpired() bool {
	return !s.ExpiresAt.After(time.Now())
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByToken(ctx context.Context, token string) (*Session, error)
	FindByUserID(ctx context.Context, userID string) ([]Session, error)
	Update(ctx context.Context, session *Session) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type SessionCache interface {
	Get(ctx context.Context, token string) (*Session, error)
	Set(ctx context.Context, session *Session, ttl time.Duration) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
