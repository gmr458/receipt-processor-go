package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/rs/xid"
)

// contextKey is an unexported type so keys stored in a request's context
// can't collide with keys set by other packages.
type contextKey int

const (
	requestIDCtxKey contextKey = iota
	loggerCtxKey
)

// requestIDFromContext returns the request ID stored by requestLogger, if any.
func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDCtxKey).(string)
	return id
}

// loggerFromContext returns a logger already tagged with this request's ID.
// Falls back to the app's base logger if called outside a request (e.g. tests),
// so callers never need a nil check.
func (app *app) loggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerCtxKey).(*slog.Logger); ok {
		return l
	}
	return app.logger
}

// requestLogger wraps every request once, timing it and capturing its final
// status/byte count, then emits a single structured access-log line. It also
// attaches a request ID to the request's context (and returns it as
// X-Request-Id), so any logging a handler — or errors.go's logError — does
// further down the chain can be correlated with this line.
func (app *app) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := xid.New().String()
		w.Header().Set("X-Request-Id", requestID)

		reqLogger := app.logger.With("request_id", requestID)
		ctx := context.WithValue(r.Context(), requestIDCtxKey, requestID)
		ctx = context.WithValue(ctx, loggerCtxKey, reqLogger)
		r = r.WithContext(ctx)

		mw := newStatusResponseWriter(w)

		ip, err := clientIP(r, app.config.limiter.trustedProxyHeader)
		if err != nil {
			ip = r.RemoteAddr
		}

		next.ServeHTTP(mw, r)

		duration := time.Since(start)
		status := mw.StatusCode()
		durationMs := float64(duration.Microseconds()) / 1000.0

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"route_pattern", r.Pattern,
			"status", status,
			"duration_ms", durationMs,
			"bytes_written", mw.bytes,
			"remote_addr", ip,
			"user_agent", r.UserAgent(),
		}

		switch {
		case status >= 500:
			reqLogger.Error("request completed", attrs...)

		case status >= 400:
			reqLogger.Warn("request completed", attrs...)

		default:
			reqLogger.Info("request completed", attrs...)
		}
	})
}
