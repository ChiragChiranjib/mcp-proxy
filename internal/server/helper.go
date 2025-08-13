package server

import (
	"encoding/json"
	"net/http"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
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

// writeRPCResult writes a JSON-RPC success response.
func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  any             `json:"result"`
	}{JSONRPC: "2.0", ID: id, Result: result})
}

// writeRPCError writes a JSON-RPC error response with a message.
func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: msg},
	})
}
