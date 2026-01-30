package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"api-gateway/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

var (
	authServiceURL   = getEnv("AUTH_SERVICE_URL", "http://auth-service:8082")
	goAPIServiceURL  = getEnv("GO_API_SERVICE_URL", "http://go-api:8080")
	pythonServiceURL = getEnv("PYTHON_SERVICE_URL", "http://python-processor:8081")
	redisAddr        = getEnv("REDIS_ADDR", "redis:6379")
	jwtSecret        = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
)

func main() {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// CORS middleware
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}

	originsList := strings.Split(allowedOrigins, ",")
	for i := range originsList {
		originsList[i] = strings.TrimSpace(originsList[i])
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   originsList,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"api-gateway"}`))
	})

	// Rate limiter middleware
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)

	// Circuit breaker for each service
	authCB := middleware.NewCircuitBreaker("auth-service")
	goAPICB := middleware.NewCircuitBreaker("go-api")
	pythonCB := middleware.NewCircuitBreaker("python-processor")

	// JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware(jwtSecret)

	// Auth service routes (no auth required for login/register)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(rateLimiter.RateLimit)
		r.Use(authCB.Middleware)
		r.HandleFunc("/*", createProxy(authServiceURL))
	})

	// Go API routes (protected)
	r.Route("/api/v1/data", func(r chi.Router) {
		r.Use(rateLimiter.RateLimit)
		r.Use(jwtMiddleware.Authenticate)
		r.Use(goAPICB.Middleware)
		r.HandleFunc("/*", createProxy(goAPIServiceURL))
	})

	// Python processor routes
	r.Route("/api/v1/process", func(r chi.Router) {
		r.Use(rateLimiter.RateLimit)
		r.Use(pythonCB.Middleware)
		r.HandleFunc("/*", createProxy(pythonServiceURL))
	})

	// Start server
	port := getEnv("PORT", "8000")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("API Gateway starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func createProxy(targetURL string) http.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Custom director to preserve original request path
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Explicitly handle OPTIONS to prevent 405 from backends that don't implementation it
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		proxy.ServeHTTP(w, r)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
