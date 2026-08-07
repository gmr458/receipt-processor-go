package main

import "net/http"

// statusResponseWriter wraps http.ResponseWriter to capture the status code
// and byte count actually written, neither of which the standard interface
// exposes. Implements Unwrap() so http.NewResponseController and interface
// assertions (Flusher, Hijacker, etc.) can still reach the underlying writer
// if a handler needs them.
type statusResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{ResponseWriter: w}
}

func (mw *statusResponseWriter) WriteHeader(statusCode int) {
	if !mw.wroteHeader {
		mw.status = statusCode
		mw.wroteHeader = true
	}
	mw.ResponseWriter.WriteHeader(statusCode)
}

func (mw *statusResponseWriter) Write(b []byte) (int, error) {
	if !mw.wroteHeader {
		mw.WriteHeader(http.StatusOK)
	}
	n, err := mw.ResponseWriter.Write(b)
	mw.bytes += n
	return n, err
}

func (mw *statusResponseWriter) Unwrap() http.ResponseWriter {
	return mw.ResponseWriter
}

// StatusCode returns the status written, defaulting to 200 if the handler
// never explicitly called WriteHeader — matching net/http's own default.
func (mw *statusResponseWriter) StatusCode() int {
	if mw.status == 0 {
		return http.StatusOK
	}
	return mw.status
}
