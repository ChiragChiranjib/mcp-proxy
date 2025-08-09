package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
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
			srv, err := buildMCPServer(
				context.Background(),
				deps,
				vsID,
			)
			if err != nil {
				deps.Logger.Error(
					"MCP_STREAMABLE_HANDLER_BUILD_ERROR",
					"error", err,
				)
				return sdk.NewServer(
					&sdk.Implementation{
						Name:    "mcp-proxy",
						Version: "1.0.0",
					},
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
	srv := sdk.NewServer(
		&sdk.Implementation{Name: "mcp-proxy", Version: "1.0.0"},
		nil,
	)

	tools, err := deps.Tools.ListForVirtualServer(ctx, vsID)
	if err != nil {
		deps.Logger.Error(
			"BUILD_MCP_SERVER_LIST_TOOLS_ERROR",
			"error", err,
		)
		return nil, fmt.Errorf("list tools: %w", err)
	}
	deps.Logger.Info(
		"BUILD_MCP_SERVER_LIST_TOOLS_OK",
		"count", len(tools),
	)
	for _, t := range tools {

		var annotations *sdk.ToolAnnotations
		if len(t.Annotations) > 0 {
			var a sdk.ToolAnnotations
			if err := json.Unmarshal(t.Annotations, &a); err != nil {
				deps.Logger.Error(
					"BUILD_MCP_SERVER_DECODE_ANN_ERROR",
					"error", err,
					"tool_mod", t.ModifiedName,
				)
			} else {
				annotations = &a
			}
		}

		// Define tool on the proxy server.
		srv.AddTool(
			&sdk.Tool{
				Name:        t.ModifiedName,
				Description: fmt.Sprintf("Proxy for %s", t.OriginalName),
				Annotations: annotations,
				InputSchema: &jsonschema.Schema{},
			},
			func(
				reqCtx context.Context,
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
						Content: []sdk.Content{
							&sdk.TextContent{Text: err.Error()},
						},
						StructuredContent: map[string]any{
							"error": err.Error(),
						},
					}, nil
				}
				// Build a client to the upstream using streamable transport
				// and auth headers.
				headers := map[string]string{}
				switch hub.AuthType {
				case m.AuthTypeBearer:
					// If encrypted JSON, decrypt first.
					token := string(hub.AuthValue)
					if deps.Encrypter != nil && len(hub.AuthValue) > 0 &&
						hub.AuthValue[0] == '{' {
						if b, derr := deps.Encrypter.DecryptFromJSON(
							hub.AuthValue,
						); derr == nil {
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
				case m.AuthTypeCustomHeaders:
					// hub.AuthValue contains JSON map of headers.
					var hdrs map[string]string
					_ = json.Unmarshal(hub.AuthValue, &hdrs)
					for k, v := range hdrs {
						headers[k] = v
					}
				}
				httpClient := httpclient.NewHTTPClient(
					httpclient.WithHeaders(headers),
				)
				baseTransport := sdk.NewStreamableClientTransport(
					hub.ServerURL,
					&sdk.StreamableClientTransportOptions{
						HTTPClient: httpClient,
					},
				)
				var transport sdk.Transport = baseTransport
				transport = sdk.NewLoggingTransport(baseTransport, os.Stdout)
				client := sdk.NewClient(
					&sdk.Implementation{
						Name:    "mcp-proxy-client",
						Version: "1.0.0",
						Title:   "mcp-gateway-client",
					},
					nil,
				)
				// Ensure upstream connect does not hang forever.
				//connectCtx, cancel := context.WithTimeout(reqCtx, 15*time.Second)
				//defer cancel()
				cs, err := client.Connect(ctx, transport)
				if err != nil {
					deps.Logger.Error(
						"PROXY_TOOL_CONNECT_ERROR",
						"error", err,
					)
					return &sdk.CallToolResultFor[any]{
						Content: []sdk.Content{
							&sdk.TextContent{Text: err.Error()},
						},
						StructuredContent: map[string]any{
							"error": err.Error(),
						},
					}, nil
				}
				defer func() { _ = cs.Close() }()

				// Forward the call using original name and the same args.
				callCtx, cCancel := context.WithTimeout(reqCtx, 60*time.Second)
				defer cCancel()
				upstreamRes, err := cs.CallTool(callCtx, &sdk.CallToolParams{
					Name:      t.ModifiedName,
					Arguments: params.Arguments,
				})
				if err != nil {
					deps.Logger.Error(
						"PROXY_TOOL_CALL_ERROR",
						"error", err,
					)
					return &sdk.CallToolResultFor[any]{
						Content: []sdk.Content{
							&sdk.TextContent{Text: err.Error()},
						},
						StructuredContent: map[string]any{
							"error": err.Error(),
						},
					}, nil
				}
				// Pass through the upstream response as-is.
				deps.Logger.Info(
					"PROXY_TOOL_OK",
					"tool_mod", t.ModifiedName,
				)
				return upstreamRes, nil
			},
		)
	}
	deps.Logger.Info("BUILD_MCP_SERVER_OK", "vs_id", vsID)
	return srv, nil
}
