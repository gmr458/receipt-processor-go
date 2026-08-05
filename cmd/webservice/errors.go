package main

import (
	"net/http"
	"runtime/debug"

	"github.com/gmr458/receipt-processor/errs"
)

var codes = map[string]int{
	errs.EINVALID:              http.StatusBadRequest,
	errs.EUNAUTHORIZED:         http.StatusUnauthorized,
	errs.EFORBIDDEN:            http.StatusForbidden,
	errs.ENOTFOUND:             http.StatusNotFound,
	errs.ENOTACCEPTABLE:        http.StatusNotAcceptable,
	errs.ECONFLICT:             http.StatusConflict,
	errs.EUNPROCESSABLECONTENT: http.StatusUnprocessableEntity,
	errs.ETOOMANYREQUESTS:      http.StatusTooManyRequests,
	errs.EINTERNAL:             http.StatusInternalServerError,
	errs.ENOTIMPLEMENTED:       http.StatusNotImplemented,
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
		"stack":  string(debug.Stack()),
	})
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	code := errs.ErrorCode(err)
	message := errs.ErrorMessage(err)
	details := errs.ErrorDetails(err)

	if code == errs.EINTERNAL {
		app.logError(r, err)
	}

	status := errorStatusCode(code)
	app.sendJSON(w, status, envelope{
		"message": message,
		"details": details,
	}, nil)
}
