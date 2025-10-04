package main

import (
	"net/http"

	"github.com/gmr458/receipt-processor/domain"
)

var codes = map[string]int{
	domain.EINVALID:              http.StatusBadRequest,
	domain.EUNAUTHORIZED:         http.StatusUnauthorized,
	domain.EFORBIDDEN:            http.StatusForbidden,
	domain.ENOTFOUND:             http.StatusNotFound,
	domain.ENOTACCEPTABLE:        http.StatusNotAcceptable,
	domain.ECONFLICT:             http.StatusConflict,
	domain.EUNPROCESSABLECONTENT: http.StatusUnprocessableEntity,
	domain.ETOOMANYREQUESTS:      http.StatusTooManyRequests,
	domain.EINTERNAL:             http.StatusInternalServerError,
	domain.ENOTIMPLEMENTED:       http.StatusNotImplemented,
}

func errorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

func (app *app) logError(r *http.Request, err error) {
	app.logger.Error(err.Error(), "details", map[string]string{
		"method": r.Method,
		"url":    r.URL.String(),
	})
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	code := domain.ErrorCode(err)
	message := domain.ErrorMessage(err)
	details := domain.ErrorDetails(err)

	if code == domain.EINTERNAL {
		app.logError(r, err)
	}

	status := errorStatusCode(code)
	app.sendJSON(w, status, envelope{
		"message": message,
		"details": details,
	}, nil)
}
