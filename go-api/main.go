package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	dbURL     = getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/microservices?sslmode=disable")
	redisAddr = getEnv("REDIS_ADDR", "redis:6379")
)

type Stats struct {
	TotalUsers    int       `json:"total_users"`
	TotalRequests int       `json:"total_requests"`
	Timestamp     time.Time `json:"timestamp"`
}

type CachedData struct {
	Message   string    `json:"message"`
	Data      string    `json:"data"`
	CachedAt  time.Time `json:"cached_at"`
	FromCache bool      `json:"from_cache"`
}

func main() {
	ctx := context.Background()

	// Initialize PostgreSQL
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis")
	}

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)



	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "go-api",
		})
	})

	// API routes
	r.Route("/api/v1/data", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handleGetData(w, r, redisClient)
		})
		
		r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
			handleGetStats(w, r, dbPool)
		})

		// Housing API routes
		r.Route("/housing", func(r chi.Router) {
			r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
				handleGetHousingData(w, r, dbPool)
			})
			r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
				handleGetHousingStats(w, r, dbPool)
			})
			r.Get("/charts", func(w http.ResponseWriter, r *http.Request) {
				handleGetHousingCharts(w, r, dbPool)
			})
		})
	})

	// Start server
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Go API Service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func handleGetData(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	ctx := context.Background()
	cacheKey := "api:data:sample"

	// Try to get from cache
	cachedValue, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		response := CachedData{
			Message:   "Data retrieved from cache",
			Data:      cachedValue,
			CachedAt:  time.Now(),
			FromCache: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Cache miss - generate new data
	newData := "Sample data generated at " + time.Now().Format(time.RFC3339)
	
	// Store in cache for 5 minutes
	_ = redisClient.Set(ctx, cacheKey, newData, 5*time.Minute).Err()

	response := CachedData{
		Message:   "Data generated and cached",
		Data:      newData,
		CachedAt:  time.Now(),
		FromCache: false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetStats(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	ctx := context.Background()

	var totalUsers int
	err := dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		log.Printf("Error querying users: %v", err)
		totalUsers = 0
	}

	stats := Stats{
		TotalUsers:    totalUsers,
		TotalRequests: 0, // Could be tracked separately
		Timestamp:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
