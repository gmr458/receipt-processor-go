package main

import (
	"net/http"
	"strconv"
	"time"
)

func (api *app) badRequest(w http.ResponseWriter, errMsg string, details map[string]string) {
	api.sendJSON(w, http.StatusBadRequest, envelope{"error": errMsg, "details": details}, nil)
}

func (api *app) tooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
	code := http.StatusTooManyRequests
	message := http.StatusText(code)
	api.sendJSON(w, code, envelope{"error": message}, nil)
}
