package main

type config struct {
	// HTTP Server's port
	port int

	// Application Environment (development|staging|production)
	env string

	// Database Config
	db struct {
		// Data Source Name
		dsn string
	}

	// Limit Rate Config
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}

	// CORS Config
	cors struct {
		// List of trusted origins, separated by spaces
		trustedOrigins []string
	}
}
