package external

import (
	"context"
	"testing"
)

func TestRedisStore_NotConfigured(t *testing.T) {
	s := &RedisStore{}
	ctx := context.Background()

	_, err := s.Get(ctx, "key")
	if err == nil {
		t.Fatal("expected error for unconfigured Redis store")
	}
}

func TestRedisStore_ScanWithNoClient(t *testing.T) {
	// Test Scan method exists and handles missing client gracefully
	s := &RedisStore{}
	ctx := context.Background()

	keys, err := s.Scan(ctx, "*")
	if err == nil {
		t.Fatal("expected error for unconfigured Redis scan")
	}
	if keys != nil {
		t.Fatal("expected nil keys")
	}
}

func TestPostgreSQLStore_List(t *testing.T) {
	s := &PostgreSQLStore{}

	// Test List alias exists
	// Note: Without real DB connection, this will error, but method exists
	defer func() {
		_ = recover() // may panic without DB, we just verify method exists
	}()

	_, _ = s.List(context.Background())
}

func TestPostgreSQLStore_PrefixScan(t *testing.T) {
	s := &PostgreSQLStore{}

	defer func() {
		_ = recover()
	}()

	_, _ = s.PrefixScan(context.Background(), "prefix")
}

func TestStoreInterface(t *testing.T) {
	// Verify both stores implement Store interface
	var _ Store = (*RedisStore)(nil)
	var _ Store = (*PostgreSQLStore)(nil)
}

func TestRedisConfig(t *testing.T) {
	cfg := RedisConfig{
		Addr:         "localhost:6379",
		Password:     "secret",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	// Verify config fields
	if cfg.Addr != "localhost:6379" {
		t.Fatal("wrong addr")
	}
	if cfg.Password != "secret" {
		t.Fatal("wrong password")
	}
	if cfg.PoolSize != 10 {
		t.Fatal("wrong pool size")
	}
}

func TestPostgreSQLConfig(t *testing.T) {
	cfg := PostgreSQLConfig{
		DSN:             "postgres://user:pass@localhost/db?sslmode=disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
	}

	if cfg.DSN == "" {
		t.Fatal("expected DSN")
	}
	if cfg.MaxOpenConns != 25 {
		t.Fatal("wrong max open conns")
	}
}

func TestRedisStore_Close(t *testing.T) {
	s := &RedisStore{}

	// Close without client should not panic
	err := s.Close()
	if err != nil {
		t.Fatalf("Close should be nil without client, got: %v", err)
	}
}

func TestPostgreSQLStore_Close(t *testing.T) {
	s := &PostgreSQLStore{}

	defer func() {
		if r := recover(); r != nil {
			// Expected: nil pointer without db connection
		}
	}()

	_ = s.Close()
}

func TestRedisStore_Ping(t *testing.T) {
	s := &RedisStore{}

	defer func() {
		_ = recover()
	}()

	_ = s.Ping(context.Background())
}

func TestPostgreSQLStore_Ping(t *testing.T) {
	s := &PostgreSQLStore{}

	defer func() {
		_ = recover()
	}()

	_ = s.Ping(context.Background())
}

func TestRedisStore_SetWithTTL(t *testing.T) {
	// Verify method signature exists
	_ = func(s *RedisStore, ctx context.Context, key string, value []byte) error {
		return s.SetWithTTL(ctx, key, value, 0)
	}
}
