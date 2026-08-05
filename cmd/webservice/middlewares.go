package main

import (
	"fmt"
	"net/http"
)

func (api *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			var err error
			switch x := rec.(type) {
			case error:
				err = x
			case string:
				err = fmt.Errorf("%s", x)
			default:
				err = fmt.Errorf("%v", x)
			}

			if r.Header.Get("Connection") == "Upgrade" {
				return
			}

			w.Header().Set("Connection", "close")
			api.errorResponse(w, r, err)
		}()

		next.ServeHTTP(w, r)
	})
}

func (api *app) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.config.limiter.enabled {
			ip, err := clientIP(r, api.config.limiter.trustedProxyHeader)
			if err != nil {
				api.errorResponse(w, r, err)
				return
			}

			allowed, retryAfter, err := api.rateLimiter.Allow(r.Context(), "ratelimit:"+ip)
			if err != nil {
				api.errorResponse(w, r, err)
				return
			}

			if !allowed {
				api.tooManyRequests(w, retryAfter)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
