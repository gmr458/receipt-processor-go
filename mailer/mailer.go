package mailer

import (
	"context"
	"fmt"

	"github.com/wneessen/go-mail"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	FromAddr string
	FromName string
}

type Mailer struct {
	client *mail.Client
	from   string
}

func New(cfg Config) (*Mailer, error) {
	opts := []mail.Option{mail.WithPort(cfg.Port)}
	if cfg.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create smtp client: %w", err)
	}

	from := cfg.FromAddr
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddr)
	}

	return &Mailer{
		client: client,
		from:   from,
	}, nil
}

func (m *Mailer) send(ctx context.Context, to string, subject string, body string) error {
	msg := mail.NewMsg()

	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (m *Mailer) SendPasswordReset(ctx context.Context, to string, link string) error {
	body := fmt.Sprintf("Reset your password by visiting the following link:\n\n%s\n", link)
	return m.send(ctx, to, "Reset your password", body)
}

func (m *Mailer) SendVerificationEmail(ctx context.Context, to string, link string) error {
	body := fmt.Sprintf("Verify your email address by visiting the following link:\n\n%s\n", link)
	return m.send(ctx, to, "Verify your email address", body)
}
