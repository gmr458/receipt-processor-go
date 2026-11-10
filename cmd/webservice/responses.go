package main

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

func (app *app) badRequest(w http.ResponseWriter, errMsg string, details map[string]string) {
	app.sendJSON(w, http.StatusBadRequest, envelope{"error": errMsg, "details": details}, nil)
}

func (app *app) tooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := max(1, int(math.Ceil(retryAfter.Seconds())))
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	code := http.StatusTooManyRequests
	message := http.StatusText(code)
	app.sendJSON(w, code, envelope{"error": message, "details": nil}, nil)
}
