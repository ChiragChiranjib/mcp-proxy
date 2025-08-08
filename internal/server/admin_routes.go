package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	appsvc "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

func addAdminRoutes(r *mux.Router, deps Deps, cfg Config) {
	// catalog list
	r.HandleFunc(cfg.AdminPrefix+"/catalog/servers", func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.List(r.Context())
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	}).Methods(http.MethodGet)
	// list tools for a virtual server
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers/{id}/tools", func(w http.ResponseWriter, r *http.Request) {
		vsID := mux.Vars(r)["id"]
		items, err := deps.Tools.ListForVirtualServer(r.Context(), vsID)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	}).Methods(http.MethodGet)

	// list all tools for user with filters
	r.HandleFunc(cfg.AdminPrefix+"/tools", func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		hubID := r.URL.Query().Get("hub_server_id")
		status := r.URL.Query().Get("status")
		q := r.URL.Query().Get("q")
		if deps.Tools != nil {
			items, err := deps.Tools.ListForUserFiltered(r.Context(), userID, hubID, status, q)
			if err != nil {
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": []any{}})
	}).Methods(http.MethodGet)

	// change tool status
	r.HandleFunc(cfg.AdminPrefix+"/tools/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Status string `json:"status"`
		}
		var body reqBody
		if !ReadJSON(w, r, &body) {
			return
		}
		if body.Status == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing status"})
			return
		}
		id := mux.Vars(r)["id"]
		if err := deps.Tools.SetStatus(r.Context(), id, body.Status); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodPatch)

	// delete (soft) a tool → mark as DEACTIVATED
	r.HandleFunc(cfg.AdminPrefix+"/tools/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := mux.Vars(r)["id"]
		if err := deps.Tools.SetStatus(r.Context(), id, string(types.StatusDeactivated)); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodDelete)

	// create virtual server
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers", func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		id, err := deps.Virtual.Create(r.Context(), userID)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
	}).Methods(http.MethodPost)

	// list virtual servers for user
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers", func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		items, err := deps.Virtual.ListForUser(r.Context(), userID)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	}).Methods(http.MethodGet)

	// replace virtual server tools
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers/{id}/tools", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var body struct {
			ToolIDs []string `json:"tool_ids"`
		}
		if !ReadJSON(w, r, &body) {
			return
		}
		if err := deps.Virtual.ReplaceTools(r.Context(), id, body.ToolIDs); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodPut)

	// set virtual server status
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		var body struct {
			Status string `json:"status"`
		}
		if !ReadJSON(w, r, &body) {
			return
		}
		if body.Status == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing status"})
			return
		}
		if err := deps.Virtual.SetStatus(r.Context(), id, body.Status); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodPatch)

	// delete virtual server
	r.HandleFunc(cfg.AdminPrefix+"/virtual-servers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := deps.Virtual.Delete(r.Context(), id); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodDelete)

	// hub servers list/add/delete/status
	// list hub servers for current user
	r.HandleFunc(cfg.AdminPrefix+"/hub/servers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		userID := GetUserID(r)
		items, err := deps.Hubs.ListForUser(r.Context(), userID)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	}).Methods(http.MethodGet)

	r.HandleFunc(cfg.AdminPrefix+"/hub/servers", func(w http.ResponseWriter, r *http.Request) {
		type req struct {
			MCPServerID  string          `json:"mcp_server_id"`
			Transport    string          `json:"transport"`
			Capabilities json.RawMessage `json:"capabilities"`
			AuthType     string          `json:"auth_type"`
			AuthValue    json.RawMessage `json:"auth_value"`
		}
		var b req
		if !ReadJSON(w, r, &b) {
			return
		}
		id := idgen.NewID()
		if err := deps.Hubs.Add(r.Context(), types.HubServer{
			ID:           id,
			UserID:       GetUserID(r),
			MCServerID:   b.MCPServerID,
			Status:       "ACTIVE",
			Transport:    b.Transport,
			Capabilities: b.Capabilities,
			AuthType:     b.AuthType,
			AuthValue:    b.AuthValue,
		}); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
	}).Methods(http.MethodPost)

	r.HandleFunc(cfg.AdminPrefix+"/hub/servers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodDelete:
			if err := deps.Hubs.Delete(r.Context(), id); err != nil {
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		case http.MethodPatch:
			type req struct {
				Status string `json:"status"`
			}
			var b req
			if !ReadJSON(w, r, &b) {
				return
			}
			if b.Status == "" {
				WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing status"})
				return
			}
			if err := deps.Hubs.SetStatus(r.Context(), id, b.Status); err != nil {
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}).Methods(http.MethodDelete, http.MethodPatch)

	// refresh tools from an upstream hub server
	r.HandleFunc(cfg.AdminPrefix+"/hub/servers/{id}/refresh", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		userID := GetUserID(r)
		ctx := r.Context()
		// Run refresh with a bounded timeout
		if _, err := appsvc.RefreshHubTools(ctx, deps.Hubs, deps.Tools, id, userID); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	}).Methods(http.MethodPost)
}
