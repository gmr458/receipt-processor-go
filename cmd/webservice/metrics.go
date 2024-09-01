package main

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "receipt_processor",
			Name:      "http_requests_received_total",
			Help:      "Number of HTTP requests received in total.",
		},
		[]string{},
	)

	httpResponsesSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "receipt_processor",
			Name:      "http_responses_sent_total",
			Help:      "Number of HTTP responses sent in total.",
		},
		[]string{},
	)

	httpResponsesSentByStatusTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "receipt_processor",
			Name:      "http_responses_sent_by_status_total",
			Help:      "Number of HTTP responses sent by status in total.",
		},
		[]string{"status_code"},
	)

	httpResponseDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "receipt_processor",
			Name:      "http_response_duration_seconds",
			Help:      "Duration of HTTP requests.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "status_code"},
	)

	memoryUsage = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "receipt_processor",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes.",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc)
		},
	)
)

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !mw.headerWritten {
		mw.statusCode = http.StatusOK
		mw.headerWritten = true
	}

	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (app *app) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestsReceivedTotal.WithLabelValues().Inc()

		mw := &metricsResponseWriter{wrapped: w}

		start := time.Now()

		next.ServeHTTP(mw, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(mw.statusCode)

		httpResponseDurationSeconds.With(prometheus.Labels{
			"method":      r.Method,
			"status_code": statusCode,
		}).Observe(duration)

		httpResponsesSentTotal.WithLabelValues().Inc()

		httpResponsesSentByStatusTotal.WithLabelValues(statusCode).Inc()
	})
}
