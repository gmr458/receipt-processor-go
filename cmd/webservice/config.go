package main

import "time"

type config struct {
	// HTTP Server's host
	host string

	// HTTP Server's port
	port int

	// HTTP Debug Server's port
	debugPort int

	// Application Environment (development|staging|production)
	env string

	// Database Config
	db struct {
		// Data Source Name
		dsn string
	}

	// Redis Config
	redis struct {
		// Redis's address
		addr string

		// Redis's password
		password string

		// Redis's db
		db int
	}

	// Limit Rate Config
	limiter struct {
		enabled            bool
		rps                float64
		burst              int
		trustedProxyHeader string
	}

	// CORS Config
	cors struct {
		// List of trusted origins, separated by spaces
		trustedOrigins []string
	}

	// Auth Config
	auth struct {
		// Name of the session cookie set on the client
		cookieName string

		// Session validity when "remember me" is not requested
		sessionDuration time.Duration

		// Session validity when "remember me" is requested
		rememberMeDuration time.Duration

		// How long password reset and email verification tokens stay valid
		verificationExpiry time.Duration

		// Fallback URL used to build password reset and email verification links
		defaultCallbackURL string
	}

	// SMTP Config
	smtp struct {
		// SMTP server host
		host string

		// SMTP server port
		port int

		// SMTP server username (empty disables authentication)
		username string

		// SMTP server password
		password string

		// From address used in sent emails
		fromAddr string

		// Display name used in sent emails
		fromName string
	}
}
