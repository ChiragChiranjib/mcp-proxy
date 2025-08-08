package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ChiragChiranjib/mcp-proxy/internal/server/httpclient"
)

func addMCPRoutes(r *mux.Router, deps Deps, cfg Config) {

	streamableHandler := sdk.NewStreamableHTTPHandler(func(req *http.Request) *sdk.Server {
		vsID := mux.Vars(req)["virtual_server_id"]
		srv, err := buildMCPServer(req.Context(), deps, vsID)
		if err != nil {
			return sdk.NewServer(&sdk.Implementation{Name: "mcp-proxy", Version: "0.1.0"}, nil)
		}
		return srv
	}, nil)

	r.Path(cfg.MCPMount).Handler(streamableHandler).Methods(http.MethodGet, http.MethodPost)
}

// buildMCPServer assembles a stateless sdk.Server with proxy tools from a virtual server id.
func buildMCPServer(ctx context.Context, deps Deps, vsID string) (*sdk.Server, error) {
	srv := sdk.NewServer(&sdk.Implementation{Name: "mcp-proxy", Version: "0.1.0"}, nil)

	tools, err := deps.Tools.ListForVirtualServer(ctx, vsID)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	for _, t := range tools {
		tool := &sdk.Tool{Name: t.ModifiedName, Description: fmt.Sprintf("Proxy for %s", t.OriginalName)}
		sdk.AddTool(srv, tool, func(_ context.Context, _ *sdk.ServerSession, params *sdk.CallToolParamsFor[map[string]any]) (*sdk.CallToolResultFor[any], error) {
			// Proxy implementation
			hub, err := deps.Hubs.GetWithURL(ctx, t.HubServerID)
			if err != nil {
				return &sdk.CallToolResultFor[any]{
					Content:           []sdk.Content{&sdk.TextContent{Text: err.Error()}},
					StructuredContent: map[string]any{"error": err.Error()},
					Meta:              nil,
				}, nil
			}
			// Build a client to the upstream using streamable transport and auth headers
			headers := map[string]string{}
			switch hub.AuthType {
			case "bearer":
				// If encrypted JSON, decrypt first
				token := string(hub.AuthValue)
				headers["Authorization"] = "Bearer " + token
			case "custom_headers":
				// hub.AuthValue contains JSON map of headers
				var m map[string]string
				_ = json.Unmarshal(hub.AuthValue, &m)
				for k, v := range m {
					headers[k] = v
				}
			}
			httpClient := httpclient.NewHTTPClient(httpclient.WithHeaders(headers))
			transport := sdk.NewStreamableClientTransport(hub.ServerURL, &sdk.StreamableClientTransportOptions{HTTPClient: httpClient})
			client := sdk.NewClient(&sdk.Implementation{Name: "mcp-proxy-proxy", Version: "0.1.0"}, nil)
			cs, err := client.Connect(ctx, transport)
			if err != nil {
				return &sdk.CallToolResultFor[any]{
					Content:           []sdk.Content{&sdk.TextContent{Text: err.Error()}},
					StructuredContent: map[string]any{"error": err.Error()},
				}, nil
			}
			defer func() { _ = cs.Close() }()

			// Forward the call using original name and the same arguments
			upstreamRes, err := cs.CallTool(ctx, &sdk.CallToolParams{
				Name:      t.OriginalName,
				Arguments: params.Arguments,
			})
			if err != nil {
				return &sdk.CallToolResultFor[any]{
					Content:           []sdk.Content{&sdk.TextContent{Text: err.Error()}},
					StructuredContent: map[string]any{"error": err.Error()},
				}, nil
			}
			// Pass through the upstream response as-is
			return &sdk.CallToolResultFor[any]{
				Content:           upstreamRes.Content,
				StructuredContent: upstreamRes.StructuredContent,
				Meta:              upstreamRes.Meta,
				IsError:           upstreamRes.IsError,
			}, nil
		})
	}
	return srv, nil
}
