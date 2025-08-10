package server

import (
	"context"
	"encoding/json"
	"net/http"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	ic "github.com/ChiragChiranjib/mcp-proxy/internal/httpclient"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	mclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// WriteJSON ...
func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ReadJSON ...
func ReadJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

// GetUserID ...
func GetUserID(r *http.Request) string {
	if v := r.Context().Value(ck.UserIDKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// GetUserRole returns the role string from context if present.
func GetUserRole(r *http.Request) string {
	if v := r.Context().Value(ck.UserRoleKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// CreateMCPTool constructs an mcp-go Tool from a DB tool record.
func CreateMCPTool(t m.MCPTool) mcp.Tool {
	tool := mcp.Tool{
		Name:        t.OriginalName,
		Description: t.Description,
	}
	if len(t.InputSchema) > 0 {
		tool.RawInputSchema = json.RawMessage(t.InputSchema)
	} else {
		tool.InputSchema = mcp.ToolInputSchema{Type: "object"}
	}
	if len(t.Annotations) > 0 {
		var ann mcp.ToolAnnotation
		if err := json.Unmarshal(t.Annotations, &ann); err == nil {
			tool.Annotations = ann
		}
	}
	return tool
}

// callUpstreamTool performs the upstream MCP tool call
// via mcp-go and returns result.
func callUpstreamTool(
	ctx context.Context,
	deps Deps,
	serverURL string,
	headers map[string]string,
	originalName string,
	args map[string]any,
) (*mcp.CallToolResult, error) {
	deps.Logger.Info(
		"CALL_UPSTREAM_TOOL_INIT",
		"tool_name", originalName,
		"headers_count", len(headers),
	)
	// Build HTTP client with headers
	httpClient := ic.NewHTTPClient(ic.WithHeaders(headers))
	trans, err := transport.NewStreamableHTTP(serverURL, transport.WithHTTPBasicClient(httpClient))
	if err != nil {
		deps.Logger.Error("CREATE_TRANSPORT_ERROR", "error", err)
		return nil, err
	}
	c := mclient.NewClient(trans)
	if err := c.Start(ctx); err != nil {
		deps.Logger.Error("UPSTREAM_CLIENT_START_ERROR", "error", err)
		return nil, err
	}
	defer func() { _ = c.Close() }()
	// Initialize
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{
		Request: mcp.Request{Method: string(mcp.MethodInitialize)},
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo:      mcp.Implementation{Name: "mcp-proxy-upstream", Version: "1.0.0"},
			Capabilities:    mcp.ClientCapabilities{},
		},
	}); err != nil {
		deps.Logger.Error("UPSTREAM_INITIALIZE_ERROR", "error", err)
		return nil, err
	}
	deps.Logger.Info("UPSTREAM_INITIALIZE_OK")
	// Call tool (use original name expected by upstream)
	res, err := c.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      originalName,
		Arguments: args,
	}})
	if err != nil {
		deps.Logger.Error("UPSTREAM_TOOL_CALL_ERROR", "error", err)
		return nil, err
	}
	deps.Logger.Info("UPSTREAM_TOOL_CALL_OK", "tool_name", originalName)
	return res, nil
}
