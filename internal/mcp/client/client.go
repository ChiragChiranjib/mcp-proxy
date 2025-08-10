package client

import (
	"context"
	"encoding/json"
	"time"

	ic "github.com/ChiragChiranjib/mcp-proxy/internal/httpclient"
	mclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConnectStreamable creates an mcp-go streamable HTTP client to the given URL.
// Caller must Close via client.Close().
func ConnectStreamable(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*mclient.Client, error) {

	httpClient := ic.NewHTTPClient(ic.WithHeaders(headers))
	trans, err := transport.NewStreamableHTTP(url,
		transport.WithHTTPBasicClient(httpClient))
	if err != nil {
		return nil, err
	}
	c := mclient.NewClient(trans)
	if err := c.Start(ctx); err != nil {
		return nil, err
	}
	// Initialize
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Request: mcp.Request{Method: string(mcp.MethodInitialize)},
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcp-proxy-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// ListTools connects and lists tools using mcp-go client.
func ListTools(
	ctx context.Context,
	url string,
	headers map[string]string) (*mcp.ListToolsResult, error) {
	c, err := ConnectStreamable(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer func() { _ = c.Close() }()
	return c.ListTools(ctx, mcp.ListToolsRequest{})
}

// CallTool connects and calls a tool with the provided name and arguments.
func CallTool(
	ctx context.Context,
	url string,
	toolName string,
	args map[string]any,
	headers map[string]string) (*mcp.CallToolResult, error) {
	c, err := ConnectStreamable(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer func() { _ = c.Close() }()
	cctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	return c.CallTool(cctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	}})
}

// InitCapabilities initializes and returns the negotiated capabilities.
func InitCapabilities(
	ctx context.Context,
	url string,
	headers map[string]string,
) ([]byte, error) {
	c, err := ConnectStreamable(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer func() { _ = c.Close() }()
	// Marshal server capabilities as raw JSON for now
	caps := c.GetServerCapabilities()
	return json.Marshal(caps)
}
