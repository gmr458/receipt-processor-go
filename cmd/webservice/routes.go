package main

import "net/http"

func (app *app) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /receipts/process", app.handlerProcessReceipts)
	mux.HandleFunc("GET /receipts/{id}/points", app.handlerGetPoints)
	mux.HandleFunc("GET /receipts", app.handlerGetReceipts)

	mux.HandleFunc("POST /api/auth/sign-up/email", app.handlerSignUpEmail)
	mux.HandleFunc("POST /api/auth/sign-in/email", app.handlerSignInEmail)
	mux.HandleFunc("POST /api/auth/sign-in/username", app.handlerSignInUsername)
	mux.HandleFunc("GET /api/auth/get-session", app.handlerGetSession)
	mux.HandleFunc("POST /api/auth/get-session", app.handlerGetSession)
	mux.HandleFunc("POST /api/auth/sign-out", app.handlerSignOut)
	mux.HandleFunc("POST /api/auth/is-username-available", app.handlerIsUsernameAvailable)
	mux.Handle("GET /api/auth/list-sessions", app.requireAuth(http.HandlerFunc(app.handlerListSessions)))
	mux.Handle("POST /api/auth/revoke-session", app.requireAuth(http.HandlerFunc(app.handlerRevokeSession)))
	mux.Handle("POST /api/auth/update-session", app.requireAuth(http.HandlerFunc(app.handlerUpdateSession)))
	mux.Handle("POST /api/auth/revoke-sessions", app.requireAuth(http.HandlerFunc(app.handlerRevokeAllSessions)))
	mux.Handle("POST /api/auth/revoke-other-sessions", app.requireAuth(http.HandlerFunc(app.handlerRevokeOtherSessions)))
	mux.Handle("POST /api/auth/update-user", app.requireAuth(http.HandlerFunc(app.handlerUpdateUser)))
	mux.Handle("POST /api/auth/change-email", app.requireAuth(http.HandlerFunc(app.handlerChangeEmail)))
	mux.Handle("POST /api/auth/verify-password", app.requireAuth(http.HandlerFunc(app.handlerVerifyPassword)))
	mux.Handle("POST /api/auth/delete-user", app.requireAuth(http.HandlerFunc(app.handlerDeleteUser)))
	mux.Handle("POST /api/auth/change-password", app.requireAuth(http.HandlerFunc(app.handlerChangePassword)))
	mux.HandleFunc("POST /api/auth/request-password-reset", app.handlerRequestPasswordReset)
	mux.HandleFunc("GET /api/auth/reset-password/{token}", app.handlerValidateResetToken)
	mux.HandleFunc("POST /api/auth/reset-password", app.handlerResetPassword)
	mux.HandleFunc("GET /api/auth/verify-email", app.handlerVerifyEmail)
	mux.Handle("POST /api/auth/send-verification-email", app.requireAuth(http.HandlerFunc(app.handlerSendVerificationEmail)))
	mux.HandleFunc("GET /api/auth/ok", app.handlerOK)
	mux.HandleFunc("GET /api/auth/error", app.handlerError)
	mux.HandleFunc("POST /api/auth/get-access-token", app.handlerGetAccessToken)
	mux.HandleFunc("POST /api/auth/refresh-token", app.handlerRefreshToken)
	mux.HandleFunc("GET /api/auth/delete-user/callback", app.handlerDeleteUserCallback)

	return app.requestLogger(app.metrics(app.recoverPanic(app.loadSession(app.corsHandler.Handler(app.rateLimit(mux))))))
}
