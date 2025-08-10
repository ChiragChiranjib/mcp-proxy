package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/mark3labs/mcp-go/mcp"
	mserver "github.com/mark3labs/mcp-go/server"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ChiragChiranjib/mcp-proxy/internal/httpclient"
	mcpclient "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/client"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

func addMCPRoutes(r *mux.Router, deps Deps, cfg Config) {
	// Keep go-sdk path compiled but unused; prefer mcp-go stateless server.
	mountMCPGoRoutes(r, deps, cfg)

	// streamableHandler := sdk.NewStreamableHTTPHandler(
	//	func(req *http.Request) *sdk.Server {
	//		vsID := mux.Vars(req)["virtual_server_id"]
	//		deps.Logger.Info(
	//			"MCP_STREAMABLE_HANDLER_INIT",
	//			"method", req.Method,
	//			"path", req.URL.Path,
	//			"vs_id", vsID,
	//		)
	//		srv, err := buildMCPServer(
	//			context.Background(),
	//			deps,
	//			vsID,
	//		)
	//		if err != nil {
	//			deps.Logger.Error(
	//				"MCP_STREAMABLE_HANDLER_BUILD_ERROR",
	//				"error", err,
	//			)
	//			return sdk.NewServer(
	//				&sdk.Implementation{
	//					Name:    "mcp-proxy",
	//					Version: "1.0.0",
	//				},
	//				nil,
	//			)
	//		}
	//		deps.Logger.Info("MCP_STREAMABLE_HANDLER_BUILD_SUCCESS")
	//		return srv
	//	},
	//	nil,
	// )
	//
	// r.Path(cfg.MCPMount).Handler(streamableHandler).
	// Methods(http.MethodGet, http.MethodPost)
}

// buildMCPServer assembles a stateless sdk.Server
// with proxy tools from a virtual server id.
func _(
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
		"BUILD_MCP_SERVER_LIST_TOOLS_SUCCESS",
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

		// Decode input schema if available
		var inputSchema *jsonschema.Schema
		if len(t.InputSchema) > 0 {
			var s jsonschema.Schema
			if err := json.Unmarshal(t.InputSchema, &s); err != nil {
				deps.Logger.Error(
					"BUILD_MCP_SERVER_DECODE_SCHEMA_ERROR",
					"error", err,
					"tool_mod", t.ModifiedName,
				)
			} else {
				inputSchema = &s
			}
		}
		if inputSchema == nil {
			inputSchema = &jsonschema.Schema{}
		}

		// Define tool on the proxy server.
		srv.AddTool(
			&sdk.Tool{
				Name:        t.ModifiedName,
				Description: fmt.Sprintf("Proxy for %s", t.OriginalName),
				Annotations: annotations,
				InputSchema: inputSchema,
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
								"PROXY_TOOL_DECRYPT_BEARER_SUCCESS",
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
					hub.URL,
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
				// connectCtx, cancel := context.WithTimeout(reqCtx, 15*time.Second)
				// defer cancel()
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
					"PROXY_TOOL_SUCCESS",
					"tool_mod", t.ModifiedName,
				)
				return upstreamRes, nil
			},
		)
	}
	deps.Logger.Info("BUILD_MCP_SERVER_SUCCESS", "vs_id", vsID)
	return srv, nil
}

// mountMCPGoRoutes mounts an mcp-go streamable
// HTTP server (stateless) at cfg.MCPMount.
func mountMCPGoRoutes(r *mux.Router, deps Deps, cfg Config) {
	// Build MCP server core
	srv := mserver.NewMCPServer(
		"mcp-proxy-server", "1.0.0",
		mserver.WithLogging(),
		mserver.WithToolCapabilities(true),
		mserver.WithHooks(&mserver.Hooks{
			OnRequestInitialization: []mserver.OnRequestInitializationFunc{
				func(ctx context.Context, _ any, _ any) error {
					// Extract virtual_server_id from request (set via context func)
					v := ctx.Value(requestVarsKey{})
					vars, _ := v.(map[string]string)
					vsID := ""
					if vars != nil {
						vsID = vars["virtual_server_id"]
					}

					// Session-scoped tools
					sess := mserver.ClientSessionFromContext(ctx)
					sTools, ok := sess.(mserver.SessionWithTools)
					if !ok {
						return nil
					}
					deps.Logger.Info("SESSION_TOOLS_LOAD_INIT", "vs_id", vsID)
					tools, err := deps.Tools.ListForVirtualServer(ctx, vsID)
					if err != nil {
						deps.Logger.Error("SESSION_TOOLS_LOAD_ERROR", "error", err)
						return nil
					}
					deps.Logger.Info("SESSION_TOOLS_LOAD_SUCCESS", "count", len(tools))

					st := make(map[string]mserver.ServerTool, len(tools))
					for _, t := range tools {
						// Build tool from DB record
						tool := CreateMCPTool(t)

						handler := makeProxyToolHandler(
							deps,
							t.OriginalName,
							t.MCPHubServerID,
						)
						st[t.OriginalName] = mserver.ServerTool{
							Tool:    tool,
							Handler: handler,
						}
					}
					sTools.SetSessionTools(st)
					deps.Logger.Info("SESSION_TOOLS_SET_SUCCESS", "count", len(st))
					return nil
				},
			},
		}),
	)

	srv.AddTools()

	deps.Logger.Info("STREAMABLE_SERVER_BUILD_INIT")
	// Build streamable HTTP handler in stateless mode and inject mux vars
	h := mserver.NewStreamableHTTPServer(
		srv,
		mserver.WithStateLess(true),
		mserver.WithHTTPContextFunc(
			func(ctx context.Context, r *http.Request) context.Context {
				return context.WithValue(ctx, requestVarsKey{}, mux.Vars(r))
			}),
	)
	deps.Logger.Info("STREAMABLE_SERVER_BUILD_SUCCESS")

	// Mount handler
	r.Path(cfg.MCPMount).Handler(h).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	deps.Logger.Info("MCP_ROUTES_MOUNTED", "mount", cfg.MCPMount)
}

// requestVarsKey is used to stash mux vars into context for mcp-go hooks.
type requestVarsKey struct{}

// makeProxyToolHandler returns an mcp-go ToolHandlerFunc that proxies to the
// upstream MCP server using our existing repo/services and encryption logic.
func makeProxyToolHandler(
	deps Deps,
	upstreamOriginalName string,
	hubID string,
) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) { //nolint:lll
	return func(
		ctx context.Context,
		req mcp.CallToolRequest,
	) (*mcp.CallToolResult, error) {
		// Fetch hub and build headers (bearer/custom)
		hub, err := deps.Hubs.GetWithURL(ctx, hubID)
		if err != nil {
			deps.Logger.Error("PROXY_TOOL_GET_HUB_ERROR", "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: err.Error(),
					},
				},
			}, nil
		}
		// Convert to models.MCPHubServer to pass to header builder
		mhub := hub.MCPHubServer
		headers := mcpclient.BuildUpstreamHeaders(deps.Logger, deps.Encrypter, &mhub)

		// Call upstream via helper (uses encryption-aware headers)
		var args map[string]any
		if req.Params.Arguments != nil {
			if m, ok := req.Params.Arguments.(map[string]any); ok {
				args = m
			}
		}

		res, err := mcpclient.CallTool(
			ctx, hub.URL, upstreamOriginalName, args, headers)
		if err != nil {
			deps.Logger.Error("PROXY_TOOL_CALL_ERROR", "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: err.Error(),
					},
				},
			}, nil
		}
		return res, nil
	}
}
