// Package external provides external storage backends for MCP servers.
// Defines a common Store interface and reference implementations for
// Redis and PostgreSQL. Both implementations use connection pooling
// and support context-based cancellation.
package external

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound    = errors.New("key not found")
	ErrNotConfigured = errors.New("store not configured")
)

// Store is the interface for external storage backends.
// All implementations must support context-based cancellation.
type Store interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a value by key.
	Set(ctx context.Context, key string, value []byte) error
	// Delete removes a key.
	Delete(ctx context.Context, key string) error
	// Keys returns all stored keys.
	Keys(ctx context.Context) ([]string, error)
	// Close releases any resources.
	Close() error
}

// RedisStore implements Store using Redis.
// Uses connection pooling with a default pool size of 10.
type RedisStore struct {
	client *redis.Client
	mu     sync.RWMutex
}

// RedisConfig configures a RedisStore.
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
}

// NewRedis creates a new Redis-backed store.
func NewRedis(cfg RedisConfig) *RedisStore {
	opts := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
	}

	rdb := redis.NewClient(opts)
	return &RedisStore{
		client: rdb,
	}
}

// Get retrieves a value by key from Redis.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	if s.client == nil {
		return nil, ErrNotConfigured
	}
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}
	return []byte(val), nil
}

// Set stores a value by key in Redis with no expiration.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte) error {
	return s.SetWithTTL(ctx, key, value, 0)
}

// SetWithTTL stores a value with a TTL (0 = no expiration).
func (s *RedisStore) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s.client == nil {
		return ErrNotConfigured
	}
	err := s.client.Set(ctx, key, string(value), ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// Delete removes a key from Redis.
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return ErrNotConfigured
	}
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}
	return nil
}

// Keys returns all keys using Redis SCAN.
func (s *RedisStore) Keys(ctx context.Context) ([]string, error) {
	return s.Scan(ctx, "*")
}

// Scan scans keys with pattern using Redis SCAN.
func (s *RedisStore) Scan(ctx context.Context, pattern string) ([]string, error) {
	if s.client == nil {
		return nil, ErrNotConfigured
	}
	var allKeys []string
	var cursor uint64

	for {
		keys, cur, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("redis scan: %w", err)
		}

		allKeys = append(allKeys, keys...)
		cursor = cur
		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

// Close closes the Redis connection.
func (s *RedisStore) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

// Ping checks Redis connectivity.
func (s *RedisStore) Ping(ctx context.Context) error {
	if s.client == nil {
		return ErrNotConfigured
	}
	return s.client.Ping(ctx).Err()
}

// PostgreSQLStore implements Store using PostgreSQL.
// Uses a table named "mcp_storage" with columns (key TEXT PRIMARY KEY, value BYTEA).
type PostgreSQLStore struct {
	db *sql.DB
	mu sync.RWMutex
}

// PostgreSQLConfig configures a PostgreSQLStore.
type PostgreSQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewPostgreSQL creates a new PostgreSQL-backed store.
func NewPostgreSQL(cfg PostgreSQLConfig) (*PostgreSQLStore, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	store := &PostgreSQLStore{db: db}
	if err := store.ensureSchema(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return store, nil
}

// ensureSchema creates the storage table if it doesn't exist.
func (s *PostgreSQLStore) ensureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS mcp_storage (
			key   TEXT PRIMARY KEY,
			value BYTEA NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
}

// Get retrieves a value by key from PostgreSQL.
func (s *PostgreSQLStore) Get(ctx context.Context, key string) ([]byte, error) {
	if s.db == nil {
		return nil, ErrNotConfigured
	}
	var value []byte
	err := s.db.QueryRowContext(ctx, "SELECT value FROM mcp_storage WHERE key = $1", key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("postgres get: %w", err)
	}
	return value, nil
}

// Set stores a value by key in PostgreSQL (upsert).
func (s *PostgreSQLStore) Set(ctx context.Context, key string, value []byte) error {
	if s.db == nil {
		return ErrNotConfigured
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO mcp_storage (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("postgres set: %w", err)
	}
	return nil
}

// Delete removes a key from PostgreSQL.
func (s *PostgreSQLStore) Delete(ctx context.Context, key string) error {
	if s.db == nil {
		return ErrNotConfigured
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM mcp_storage WHERE key = $1", key)
	if err != nil {
		return fmt.Errorf("postgres delete: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Keys returns all stored keys from PostgreSQL.
func (s *PostgreSQLStore) Keys(ctx context.Context) ([]string, error) {
	if s.db == nil {
		return nil, ErrNotConfigured
	}
	rows, err := s.db.QueryContext(ctx, "SELECT key FROM mcp_storage")
	if err != nil {
		return nil, fmt.Errorf("postgres keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("scan key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return keys, nil
}

// List is an alias for Keys to maintain interface compatibility.
func (s *PostgreSQLStore) List(ctx context.Context) ([]string, error) {
	return s.Keys(ctx)
}

// Close closes the PostgreSQL connection.
func (s *PostgreSQLStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Ping checks PostgreSQL connectivity.
func (s *PostgreSQLStore) Ping(ctx context.Context) error {
	if s.db == nil {
		return ErrNotConfigured
	}
	return s.db.PingContext(ctx)
}

// PrefixScan returns all keys with the given prefix.
func (s *PostgreSQLStore) PrefixScan(ctx context.Context, prefix string) ([]string, error) {
	if s.db == nil {
		return nil, ErrNotConfigured
	}
	rows, err := s.db.QueryContext(ctx, "SELECT key FROM mcp_storage WHERE key LIKE $1", prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("postgres prefix scan: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("scan key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}
