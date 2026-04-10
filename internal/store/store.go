package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const cacheTTL = 24 * time.Hour

type Store struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func New(db *pgxpool.Pool, rdb *redis.Client) *Store {
	return &Store{db: db, rdb: rdb}
}

func (s *Store) Init(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	return err
}

func (s *Store) Save(ctx context.Context, originalURL string) (string, error) {
	code := randomCode(7)
	_, err := s.db.Exec(ctx,
		"INSERT INTO urls (code, original_url) VALUES ($1, $2)",
		code, originalURL,
	)
	if err != nil {
		return "", fmt.Errorf("failed to save url: %w", err)
	}
	// Cache it immediately
	s.rdb.Set(ctx, code, originalURL, cacheTTL)
	return code, nil
}

func (s *Store) Get(ctx context.Context, code string) (string, error) {
	// Check cache first
	val, err := s.rdb.Get(ctx, code).Result()
	if err == nil {
		return val, nil
	}

	// Fall back to Postgres
	var originalURL string
	err = s.db.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE code = $1", code,
	).Scan(&originalURL)
	if err != nil {
		return "", errors.New("url not found")
	}

	// Repopulate cache
	s.rdb.Set(ctx, code, originalURL, cacheTTL)
	return originalURL, nil
}

func randomCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
