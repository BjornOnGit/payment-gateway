package util

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total HTTP requests processed, labeled by method and path.",
		},
		[]string{"method", "path"},
	)

	TransactionsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "transactions_created_total",
			Help: "Total transactions created",
		},
	)

	SettlementsSucceededTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "settlements_succeeded_total",
			Help: "Total successful settlements",
		},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal, TransactionsCreatedTotal, SettlementsSucceededTotal)
}

// MetricsMiddleware increments RequestsTotal for each request.
// Use before handlers are executed so path/method are available.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
		next.ServeHTTP(w, r)
	})
}

// MetricsHandler returns the standard promhttp handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
