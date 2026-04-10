package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/0lawale/devops-project2/internal/handler"
	"github.com/0lawale/devops-project2/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Postgres connection
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("failed to close redis connection: %v", err)
		}
	}()

	// Initialise store and handler
	s := store.New(pool, rdb)
	if err := s.Init(ctx); err != nil {
		log.Fatalf("failed to initialise store: %v", err)
	}

	h := handler.New(s)

	r := http.NewServeMux()
	r.HandleFunc("GET /health", h.Health)
	r.HandleFunc("POST /shorten", h.Shorten)
	r.HandleFunc("GET /{code}", h.Redirect)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
