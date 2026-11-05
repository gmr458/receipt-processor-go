package auth

import (
	"context"
	"time"
)

type User struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	EmailVerified    bool       `json:"emailVerified"`
	Image            *string    `json:"image"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	Role             *string    `json:"role"`
	Banned           bool       `json:"banned"`
	BanReason        *string    `json:"banReason"`
	BanExpires       *time.Time `json:"banExpires"`
	Username         *string    `json:"username"`
	DisplayUsername  *string    `json:"displayUsername"`
	TwoFactorEnabled bool       `json:"twoFactorEnabled"`
}

type UserRepository interface {
	CreateWithCredential(ctx context.Context, user *User, account *Account, session *Session) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindById(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateEmail(ctx context.Context, id string, email string) error
	SetEmailVerified(ctx context.Context, id string, verified bool) error
	DeleteById(ctx context.Context, id string) error
}
