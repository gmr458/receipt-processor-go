package main

import (
	"net/http"

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/errs"
)

func (app *app) handlerSignUpEmail(w http.ResponseWriter, r *http.Request) {
	var input auth.SignUpDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	ip, userAgent := ipAndUserAgent(r, app)

	user, session, err := app.authService.SignUpEmail(r.Context(), input, ip, userAgent)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.setSessionCookie(w, session.Token, session.ExpiresAt)

	app.sendJSON(w, http.StatusOK, envelope{
		"token": session.Token,
		"user":  user,
	}, nil)
}

func (app *app) handlerSignInEmail(w http.ResponseWriter, r *http.Request) {
	var input auth.SignInDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	ip, userAgent := ipAndUserAgent(r, app)

	user, session, err := app.authService.SignInEmail(r.Context(), input, ip, userAgent)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.setSessionCookie(w, session.Token, session.ExpiresAt)

	app.sendJSON(w, http.StatusOK, envelope{
		"redirect": false,
		"token":    session.Token,
		"user":     user,
	}, nil)
}

func (app *app) handlerSignInUsername(w http.ResponseWriter, r *http.Request) {
	var input auth.SignInUsernameDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	ip, userAgent := ipAndUserAgent(r, app)

	user, session, err := app.authService.SignInUsername(r.Context(), input, ip, userAgent)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.setSessionCookie(w, session.Token, session.ExpiresAt)

	app.sendJSON(w, http.StatusOK, envelope{
		"redirect": false,
		"token":    session.Token,
		"user":     user,
	}, nil)
}

func (app *app) handlerGetSession(w http.ResponseWriter, r *http.Request) {
	user := authUserFromContext(r.Context())
	session := authSessionFromContext(r.Context())

	if user == nil || session == nil {
		app.sendJSON(w, http.StatusOK, nil, nil)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"session": session,
		"user":    user,
	}, nil)
}

func (app *app) handlerSignOut(w http.ResponseWriter, r *http.Request) {
	token := sessionTokenFromRequest(r, app.config.auth.cookieName)

	err := app.authService.SignOut(r.Context(), token)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.clearSessionCookie(w)

	app.sendJSON(w, http.StatusOK, envelope{
		"success": true,
	}, nil)
}

func (app *app) handlerListSessions(w http.ResponseWriter, r *http.Request) {
	user := authUserFromContext(r.Context())

	sessions, err := app.authService.ListSessions(r.Context(), user.ID)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, sessions, nil)
}

func (app *app) handlerRevokeSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	err = app.authService.RevokeSession(r.Context(), user.ID, input.Token)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
	}, nil)
}

func (app *app) handlerUpdateSession(w http.ResponseWriter, r *http.Request) {
	var input auth.UpdateSessionDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	session := authSessionFromContext(r.Context())

	updatedSession, err := app.authService.UpdateSession(r.Context(), session.Token, input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"session": updatedSession,
	}, nil)
}

func (app *app) handlerRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := authUserFromContext(r.Context())

	err := app.authService.RevokeAllSessions(r.Context(), user.ID)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.clearSessionCookie(w)

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
	}, nil)
}

func (app *app) handlerRevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	user := authUserFromContext(r.Context())
	session := authSessionFromContext(r.Context())

	err := app.authService.RevokeOtherSessions(r.Context(), user.ID, session.Token)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
	}, nil)
}

func (app *app) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	var input auth.UpdateUserDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	updatedUser, err := app.authService.UpdateUser(r.Context(), user.ID, input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"user": updatedUser,
	}, nil)
}

func (app *app) handlerChangeEmail(w http.ResponseWriter, r *http.Request) {
	var input auth.ChangeEmailDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	updatedUser, err := app.authService.ChangeEmail(r.Context(), user.ID, input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"message": "Email updated",
		"status":  true,
		"user":    updatedUser,
	}, nil)
}

func (app *app) handlerVerifyPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.VerifyPasswordDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	valid, err := app.authService.VerifyPassword(r.Context(), user.ID, input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": valid,
	}, nil)
}

func (app *app) handlerIsUsernameAvailable(w http.ResponseWriter, r *http.Request) {
	var input auth.UsernameAvailableDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	available, err := app.authService.IsUsernameAvailable(r.Context(), input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"available": available,
	}, nil)
}

func (app *app) handlerOK(w http.ResponseWriter, r *http.Request) {
	app.sendJSON(w, http.StatusOK, envelope{
		"ok": true,
	}, nil)
}

func (app *app) handlerError(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, errs.Errorf(errs.EINTERNAL, "something went wrong"))
}

func (app *app) handlerGetAccessToken(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &errs.Error{
		Code:    errs.ENOTIMPLEMENTED,
		Message: "OAuth access tokens are not implemented",
	})
}

func (app *app) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &errs.Error{
		Code:    errs.ENOTIMPLEMENTED,
		Message: "OAuth token refresh is not implemented",
	})
}

func (app *app) handlerDeleteUserCallback(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, &errs.Error{
		Code:    errs.ENOTIMPLEMENTED,
		Message: "Email-confirmed account deletion is not implemented",
	})
}

func (app *app) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password *string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	err = app.authService.DeleteUser(r.Context(), user.ID, input.Password)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.clearSessionCookie(w)

	app.sendJSON(w, http.StatusOK, envelope{
		"success": true,
		"message": "User deleted",
	}, nil)
}

func ipAndUserAgent(r *http.Request, app *app) (string, string) {
	ip, err := clientIP(r, app.config.limiter.trustedProxyHeader)
	if err != nil {
		ip = ""
	}
	return ip, r.UserAgent()
}

func (app *app) handlerChangePassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ChangePasswordDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())
	session := authSessionFromContext(r.Context())

	currentToken := ""
	if session != nil {
		currentToken = session.Token
	}

	updatedUser, err := app.authService.ChangePassword(r.Context(), user.ID, currentToken, input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"user": updatedUser,
	}, nil)
}

func (app *app) handlerRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input auth.RequestPasswordResetDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	err = app.authService.RequestPasswordReset(r.Context(), input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"message": "If an account with that email exists, a password reset email has been sent.",
		"status":  true,
	}, nil)
}

func (app *app) handlerValidateResetToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	valid, err := app.authService.ValidateResetToken(r.Context(), token)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	if !valid {
		app.errorResponse(w, r, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid or expired token",
		})
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"token": token,
	}, nil)
}

func (app *app) handlerResetPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ResetPasswordDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	err = app.authService.ResetPassword(r.Context(), input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
	}, nil)
}

func (app *app) handlerVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	user, err := app.authService.VerifyEmail(r.Context(), token)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
		"user":   user,
	}, nil)
}

func (app *app) handlerSendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CallbackURL string `json:"callbackURL"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	user := authUserFromContext(r.Context())

	err = app.authService.SendVerificationEmail(r.Context(), user.ID, input.CallbackURL)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"status": true,
	}, nil)
}
