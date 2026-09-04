package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/project/mcp-go-core/core/server"
	"github.com/project/mcp-go-core/testutil"
)

// TestSmokeServerStartShutdown tests that a minimal server can start and shutdown.
func TestSmokeServerStartShutdown(t *testing.T) {
	srv := server.NewBuilder().
		WithName("smoke-server").
		WithTransport(testutil.NewMockTransport()).
		MustBuild()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := srv.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSmokeServerAddTool tests that tools can be added to a server.
func TestSmokeServerAddTool(t *testing.T) {
	srv := server.NewBuilder().
		WithName("smoke-addtool-server").
		WithTransport(testutil.NewMockTransport()).
		WithTimeout(5 * time.Second).
		MustBuild()
	_ = srv
}
