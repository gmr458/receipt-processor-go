package auth

import "context"

type Mailer interface {
	SendPasswordReset(ctx context.Context, to string, link string) error
	SendVerificationEmail(ctx context.Context, to string, link string) error
}
