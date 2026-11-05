package auth

import (
	"context"
	"time"
)

const CredentialProviderID = "credential"

type Account struct {
	ID                    string     `json:"id"`
	Issuer                *string    `json:"issuer"`
	AccountID             string     `json:"accountId"`
	ProviderID            string     `json:"providerId"`
	UserID                string     `json:"userId"`
	AccessToken           *string    `json:"accessToken"`
	RefreshToken          *string    `json:"refreshToken"`
	IDToken               *string    `json:"idToken"`
	AccessTokenExpiresAt  *time.Time `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt *time.Time `json:"refreshTokenExpiresAt"`
	Scope                 *string    `json:"scope"`
	Password              *string    `json:"-"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type AccountRepository interface {
	FindCredentialByEmail(ctx context.Context, email string) (*Account, *User, error)
	UpdatePasswordByUserID(ctx context.Context, userID string, passwordHash string) error
}
