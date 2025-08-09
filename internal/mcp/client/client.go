package client

import (
	"context"

	"github.com/ChiragChiranjib/mcp-proxy/internal/server/httpclient"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConnectStreamable creates an MCP client using streamable HTTP transport to the given URL.
// Caller must Close the returned ClientSession.
func ConnectStreamable(ctx context.Context, url string) (*sdk.ClientSession, error) {
	httpClient := httpclient.NewHTTPClient()
	transport := sdk.NewStreamableClientTransport(
		url, &sdk.StreamableClientTransportOptions{HTTPClient: httpClient},
	)
	c := sdk.NewClient("mcp-client", "0.1.0", nil)
	return c.Connect(ctx, transport)
}

// ListTools connects to the server and returns its tools list. The caller does not need to manage session lifecycle.
func ListTools(ctx context.Context, url string) (*sdk.ListToolsResult, error) {
	cs, err := ConnectStreamable(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cs.Close() }()
	return cs.ListTools(ctx, &sdk.ListToolsParams{})
}

// CallTool connects to the server and calls a tool with the provided name and arguments.
// args should be a JSON-serializable map.
func CallTool(ctx context.Context, url string, toolName string, args map[string]any) (*sdk.CallToolResultFor[any], error) {
	cs, err := ConnectStreamable(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cs.Close() }()
	return cs.CallTool(ctx, &sdk.CallToolParams{Name: toolName, Arguments: args})
}

// InitCapabilities initializes the connection and returns raw capabilities JSON if available.
// For now, return an empty JSON object as a placeholder; extend when SDK exposes capabilities.
func InitCapabilities(ctx context.Context, url string) ([]byte, error) {
	cs, err := ConnectStreamable(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cs.Close() }()
	return []byte("{}"), nil
}
