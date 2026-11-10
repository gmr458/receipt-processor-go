package main

import (
	"net/http"
	"runtime/debug"

	"github.com/gmr458/receipt-processor-go/errs"
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
	app.loggerFromContext(r.Context()).Error(err.Error(), "details", map[string]string{
		"method": r.Method,
		"url":    r.URL.String(),
		"stack":  string(debug.Stack()),
	})
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, err error) {
	code := errs.ErrorCode(err)
	if code == errs.EINTERNAL {
		app.logError(r, err)
		app.sendJSON(w, errorStatusCode(code), envelope{
			"error":   "internal error",
			"details": nil,
		}, nil)
		return
	}

	app.sendJSON(w, errorStatusCode(code), envelope{
		"error":   errs.ErrorMessage(err),
		"details": errs.ErrorDetails(err),
	}, nil)
}
