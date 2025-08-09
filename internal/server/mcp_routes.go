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

	streamableHandler := sdk.NewStreamableHTTPHandler(
		func(req *http.Request) *sdk.Server {
			vsID := mux.Vars(req)["virtual_server_id"]
			deps.Logger.Info(
				"MCP_STREAMABLE_HANDLER_INIT",
				"method", req.Method,
				"path", req.URL.Path,
				"vs_id", vsID,
			)
			srv, err := buildMCPServer(req.Context(), deps, vsID)
			if err != nil {
				deps.Logger.Error(
					"MCP_STREAMABLE_HANDLER_BUILD_ERROR",
					"error", err,
				)
				return sdk.NewServer(
					"mcp-proxy", "0.1.0",
					nil,
				)
			}
			deps.Logger.Info("MCP_STREAMABLE_HANDLER_BUILD_OK")
			return srv
		},
		nil,
	)

	r.Path(cfg.MCPMount).Handler(streamableHandler).Methods(http.MethodGet, http.MethodPost)
}

// buildMCPServer assembles a stateless sdk.Server with proxy tools from a virtual server id.
func buildMCPServer(
	ctx context.Context,
	deps Deps,
	vsID string,
) (*sdk.Server, error) {
	deps.Logger.Info("BUILD_MCP_SERVER_INIT", "vs_id", vsID)
	srv := sdk.NewServer("mcp-proxy-22", "0.1.0", nil)

	tools, err := deps.Tools.ListForVirtualServer(ctx, vsID)
	if err != nil {
		deps.Logger.Error("BUILD_MCP_SERVER_LIST_TOOLS_ERROR", "error", err)
		return nil, fmt.Errorf("list tools: %w", err)
	}
	deps.Logger.Info("BUILD_MCP_SERVER_LIST_TOOLS_OK", "count", len(tools))
	for _, t := range tools {
		sdk.NewServerTool(t.ModifiedName, fmt.Sprintf("Proxy for %s", t.OriginalName), func(
			_ context.Context,
			_ *sdk.ServerSession,
			params *sdk.CallToolParamsFor[map[string]any],
		) (*sdk.CallToolResultFor[any], error) {
			// Proxy implementation
			deps.Logger.Info(
				"PROXY_TOOL_INIT",
				"vs_id", vsID,
				"tool_mod", t.ModifiedName,
				"hub_id", t.MCPHubServerID,
			)
			hub, err := deps.Hubs.GetWithURL(ctx, t.MCPHubServerID)
			if err != nil {
				deps.Logger.Error(
					"PROXY_TOOL_GET_HUB_ERROR",
					"error", err,
				)
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
				if deps.Encrypter != nil && len(hub.AuthValue) > 0 &&
					hub.AuthValue[0] == '{' {
					if b, derr := deps.Encrypter.DecryptFromJSON(hub.AuthValue); derr == nil {
						token = string(b)
						deps.Logger.Info(
							"PROXY_TOOL_DECRYPT_BEARER_OK",
							"len", len(token),
						)
					} else {
						deps.Logger.Error(
							"PROXY_TOOL_DECRYPT_BEARER_ERROR",
							"error", derr,
						)
					}
				}
				headers["Authorization"] = "Bearer " + token
			case "custom_headers":
				// hub.AuthValue contains JSON map of headers
				var m map[string]string
				_ = json.Unmarshal(hub.AuthValue, &m)
				for k, v := range m {
					headers[k] = v
				}
			}
			httpClient := httpclient.NewHTTPClient(
				httpclient.WithHeaders(headers),
			)
			transport := sdk.NewStreamableClientTransport(
				hub.ServerURL,
				&sdk.StreamableClientTransportOptions{HTTPClient: httpClient},
			)
			client := sdk.NewClient("mcp-proxy-proxy", "0.1.0", nil)
			cs, err := client.Connect(ctx, transport)
			if err != nil {
				deps.Logger.Error("PROXY_TOOL_CONNECT_ERROR", "error", err)
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
				deps.Logger.Error("PROXY_TOOL_CALL_ERROR", "error", err)
				return &sdk.CallToolResultFor[any]{
					Content:           []sdk.Content{&sdk.TextContent{Text: err.Error()}},
					StructuredContent: map[string]any{"error": err.Error()},
				}, nil
			}
			// Pass through the upstream response as-is
			deps.Logger.Info("PROXY_TOOL_OK", "tool_mod", t.ModifiedName)
			return &sdk.CallToolResultFor[any]{
				Content:           upstreamRes.Content,
				StructuredContent: upstreamRes.StructuredContent,
				Meta:              upstreamRes.Meta,
				IsError:           upstreamRes.IsError,
			}, nil
		})
	}
	deps.Logger.Info("BUILD_MCP_SERVER_OK", "vs_id", vsID)
	return srv, nil
}
