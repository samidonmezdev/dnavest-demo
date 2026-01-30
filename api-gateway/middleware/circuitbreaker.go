package middleware

import (
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	name    string
}

func NewCircuitBreaker(serviceName string) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
			// log.Printf("Circuit breaker '%s' changed from %s to %s", name, from, to)
		},
	}

	return &CircuitBreaker{
		breaker: gobreaker.NewCircuitBreaker(settings),
		name:    serviceName,
	}
}

func (cb *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := cb.breaker.Execute(func() (interface{}, error) {
			// Create a custom response writer to capture status code
			crw := &customResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(crw, r)
			
			// If status code indicates failure, return error
			if crw.statusCode >= 500 {
				return nil, http.ErrAbortHandler
			}
			return nil, nil
		})

		if err != nil {
			if err == gobreaker.ErrOpenState {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":"service unavailable","message":"circuit breaker is open"}`))
				return
			}
		}
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}
