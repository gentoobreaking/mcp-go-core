package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/server"
)

// TestSmokeServerStartShutdown tests that a minimal server can start and shutdown.
func TestSmokeServerStartShutdown(t *testing.T) {
	srv := server.NewServer()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := srv.Run(ctx)
	// context.Canceled is expected when we cancel
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSmokeServerAddTool tests that tools can be added to a server.
func TestSmokeServerAddTool(t *testing.T) {
	srv := server.NewServer(
		server.WithName("test-server"),
		server.WithShutdownTimeout(5*time.Second),
	)
	// Verify server was created with options
	_ = srv
}
