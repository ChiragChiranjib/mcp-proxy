package client

import (
	"context"
	"os"
	"testing"
)

// TestListTools_E2E hits the external E2E MCP server and verifies we get a
// non-nil result without error. It is skipped unless RUN_E2E_MCP=1 is set.
func TestListTools_E2E(t *testing.T) {
	if os.Getenv("RUN_E2E_MCP") != "1" {
		t.Skip("set RUN_E2E_MCP=1 to run this test")
	}

	ctx := context.Background()

	url := "https://e2e-mcp-server.dev.razorpay.in/mcp"
	res, err := ListTools(ctx, url)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
}
