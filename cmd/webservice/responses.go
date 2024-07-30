package main

import (
	"net/http"
)

func (api *app) badRequest(w http.ResponseWriter, errMsg string, details map[string]string) {
	api.sendJSON(w, http.StatusBadRequest, envelope{"error": errMsg, "details": details}, nil)
}

func (api *app) tooManyRequests(w http.ResponseWriter) {
	code := http.StatusTooManyRequests
	message := http.StatusText(code)
	api.sendJSON(w, code, envelope{"error": message}, nil)
}
