package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/errs"
)

const (
	authUserCtxKey contextKey = iota + 32
	authSessionCtxKey
)

var (
	defaultSessionCookieName = "better-auth.session_token"
	defaultSessionHours      = 7 * 24
	defaultRememberMeHours   = 30 * 24
)

// sessionTokenFromRequest returns the session token from the session cookie,
// falling back to the Authorization header's bearer token.
func sessionTokenFromRequest(r *http.Request, cookieName string) string {
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		return c.Value
	}

	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}

	return ""
}

func (app *app) newSessionCookie(token string, expires time.Time) http.Cookie {
	return http.Cookie{
		Name:     app.config.auth.cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   app.config.env == "production",
	}
}

func (app *app) setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	cookie := app.newSessionCookie(token, expires)
	http.SetCookie(w, &cookie)
}

func (app *app) clearSessionCookie(w http.ResponseWriter) {
	cookie := app.newSessionCookie("", time.Unix(0, 0))
	cookie.MaxAge = -1
	http.SetCookie(w, &cookie)
}

func authUserFromContext(ctx context.Context) *auth.User {
	user, _ := ctx.Value(authUserCtxKey).(*auth.User)
	return user
}

func authSessionFromContext(ctx context.Context) *auth.Session {
	session, _ := ctx.Value(authSessionCtxKey).(*auth.Session)
	return session
}

// loadSession resolves the request's session token (if any) and stores the
// authenticated user and session in the request context. Missing or invalid
// tokens are not an error; downstream handlers decide how to react.
func (app *app) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := sessionTokenFromRequest(r, app.config.auth.cookieName)
		if token != "" {
			user, session, err := app.authService.GetSession(r.Context(), token)
			if err != nil {
				app.errorResponse(w, r, err)
				return
			}
			if user != nil && session != nil {
				ctx := context.WithValue(r.Context(), authUserCtxKey, user)
				ctx = context.WithValue(ctx, authSessionCtxKey, session)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// requireAuth rejects requests without a valid session. Must run after
// loadSession.
func (app *app) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authUserFromContext(r.Context()) == nil {
			app.errorResponse(w, r, &errs.Error{
				Code:    errs.EUNAUTHORIZED,
				Message: "You are not authenticated",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authConfig returns the auth.Config used to build the auth service from the
// application's loaded config.
func authConfig(cfg config) auth.Config {
	return auth.Config{
		SessionDuration:    cfg.auth.sessionDuration,
		RememberMeDuration: cfg.auth.rememberMeDuration,
		VerificationExpiry: cfg.auth.verificationExpiry,
		DefaultCallbackURL: cfg.auth.defaultCallbackURL,
	}
}
