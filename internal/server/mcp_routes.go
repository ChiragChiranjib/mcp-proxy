package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mark3labs/mcp-go/mcp"
	mserver "github.com/mark3labs/mcp-go/server"

	mcpclient "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/client"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

func addMCPRoutes(r *mux.Router, deps Deps, cfg Config) {
	// Build MCP server core
	srv := mserver.NewMCPServer(
		"mcp-proxy-server", "1.0.0",
		mserver.WithLogging(),
		mserver.WithToolCapabilities(true),
	)

	deps.Logger.Info("STREAMABLE_SERVER_BUILD_INIT")
	// Build streamable HTTP handler and inject mux vars
	core := mserver.NewStreamableHTTPServer(
		srv,
		mserver.WithStateLess(true),
		mserver.WithHTTPContextFunc(
			func(ctx context.Context, r *http.Request) context.Context {
				return context.WithValue(ctx, requestVarsKey{}, mux.Vars(r))
			}),
	)
	deps.Logger.Info("STREAMABLE_SERVER_BUILD_SUCCESS")

	// Wrap core with proxy that intercepts tools/list and tools/call only.
	proxy := &proxyHTTPHandler{core: core, deps: deps}

	// Mount handler
	r.Path(cfg.MCPMount).Handler(proxy).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	deps.Logger.Info("MCP_ROUTES_MOUNTED", "mount", cfg.MCPMount)
}

// requestVarsKey is used to stash mux vars into context for mcp-go hooks.
type requestVarsKey struct{}

// proxyHTTPHandler intercepts only list_tools and call_tool POST requests,
// delegating all other methods to the core streamable handler.
type proxyHTTPHandler struct {
	core http.Handler
	deps Deps
}

func (p *proxyHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		p.core.ServeHTTP(w, r)
		return
	}

	// Enforce JSON content type similar to mcp-go
	contentType := r.Header.Get("Content-Type")
	mediaType, _, cterr := mime.ParseMediaType(contentType)
	if cterr != nil || mediaType != "application/json" {
		http.Error(w,
			"Invalid content type: must be 'application/json'",
			http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeRPCError(w, nil, mcp.PARSE_ERROR, "read body error")
		return
	}
	_ = r.Body.Close()

	// Parse minimal JSON-RPC envelope as mcp-go does
	var base struct {
		ID     json.RawMessage `json:"id"`
		Result json.RawMessage `json:"result,omitempty"`
		Error  json.RawMessage `json:"error,omitempty"`
		Method mcp.MCPMethod   `json:"method,omitempty"`
	}
	if err := json.Unmarshal(body, &base); err != nil {
		writeRPCError(w, nil, mcp.PARSE_ERROR, "request body is not valid json")
		return
	}

	switch base.Method {
	case mcp.MethodToolsList:
		p.handleListTools(w, r, base.ID)
		return
	case mcp.MethodToolsCall:
		p.handleCallTool(w, r, base.ID, body)
		return
	default:
		// Delegate to core for all other methods
		r.Body = io.NopCloser(bytes.NewReader(body))
		p.core.ServeHTTP(w, r)
		return
	}
}

func (p *proxyHTTPHandler) handleListTools(
	w http.ResponseWriter,
	r *http.Request,
	id json.RawMessage,
) {
	vars := mux.Vars(r)
	vsID := vars["virtual_server_id"]

	items, err := p.deps.Tools.ListForVirtualServer(r.Context(), vsID)
	if err != nil {
		writeRPCError(w, id, mcp.INTERNAL_ERROR, err.Error())
		return
	}

	tools := make([]mcp.Tool, 0, len(items))
	for _, t := range items {
		tool := CreateMCPTool(t)
		// Use original name for listing
		tool.Name = t.OriginalName
		tools = append(tools, tool)
	}
	res := mcp.ListToolsResult{Tools: tools}
	writeRPCResult(w, id, res)
}

func (p *proxyHTTPHandler) handleCallTool(
	w http.ResponseWriter,
	r *http.Request,
	id json.RawMessage,
	body []byte,
) {
	vars := mux.Vars(r)
	vsID := vars["virtual_server_id"]

	var req mcp.CallToolRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeRPCError(w, id, mcp.INVALID_REQUEST, "invalid call_tool")
		return
	}
	name := req.Params.Name // original name expected

	// Fetch tools for VS and find by original name
	items, err := p.deps.Tools.ListForVirtualServer(r.Context(), vsID)
	if err != nil {
		writeRPCError(w, id, mcp.INTERNAL_ERROR, err.Error())
		return
	}
	var found *m.MCPTool
	for i := range items {
		if items[i].OriginalName == name {
			found = &items[i]
			break
		}
	}
	if found == nil {
		writeRPCError(w, id, mcp.RESOURCE_NOT_FOUND, "tool not found")
		return
	}

	// Authorization: VS owner must have this server in their hub
	vs, err := p.deps.Virtual.GetByID(r.Context(), vsID)
	if err != nil {
		writeRPCError(w, id, mcp.INVALID_REQUEST, "virtual server not found")
		return
	}
	hub, err := p.deps.Hubs.GetByServerAndUser(
		r.Context(), found.MCPServerID, vs.UserID,
	)
	if err != nil {
		writeRPCError(w, id, mcp.INVALID_PARAMS, "unauthorized")
		return
	}

	headers := mcpclient.BuildUpstreamHeaders(
		p.deps.Logger, p.deps.Encrypter, &hub.MCPHubServer,
	)

	// Extract arguments
	args := map[string]any{}
	if req.Params.Arguments != nil {
		if m, ok := req.Params.Arguments.(map[string]any); ok {
			args = m
		}
	}

	res, err := mcpclient.CallTool(
		r.Context(), hub.URL, found.OriginalName, args, headers,
	)
	if err != nil {
		writeRPCError(w, id, mcp.INTERNAL_ERROR, err.Error())
		return
	}
	writeRPCResult(w, id, res)
}
