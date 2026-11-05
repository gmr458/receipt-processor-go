package auth

import (
	"context"
	"time"
)

const (
	resetPasswordIdentifierPrefix     = "reset-password:"
	emailVerificationIdentifierPrefix = "verify-email:"
)

type Verification struct {
	ID         string    `json:"id"`
	Identifier string    `json:"identifier"`
	Value      string    `json:"value"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (v *Verification) IsExpired() bool {
	return !v.ExpiresAt.After(time.Now())
}

type VerificationRepository interface {
	Create(ctx context.Context, verification *Verification) error
	FindByValue(ctx context.Context, value string) (*Verification, error)
	DeleteByIdentifier(ctx context.Context, identifier string) error
}

func resetPasswordIdentifier(userID string) string {
	return resetPasswordIdentifierPrefix + userID
}

func emailVerificationIdentifier(userID string) string {
	return emailVerificationIdentifierPrefix + userID
}
