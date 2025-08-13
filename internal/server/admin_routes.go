package server

import (
	"encoding/json"
	"net/http"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	"github.com/gorilla/mux"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	orchestrator "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub_orchestrator"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

func addAdminRoutes(r *mux.Router, deps Deps, cfg Config) {
	addCatalogRoutes(r, deps, cfg)
	addToolsRoutes(r, deps, cfg)
	addVirtualServerRoutes(r, deps, cfg)
	addHubRoutes(r, deps, cfg)
}

// Catalog routes
func addCatalogRoutes(r *mux.Router, deps Deps, cfg Config) {
	r.HandleFunc(
		cfg.AdminPrefix+"/catalog/servers",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("LIST_CATALOG_SERVERS_INIT",
				"method", r.Method,
				"path", r.URL.Path,
			)

			items, err := deps.Catalog.List(r.Context())
			if err != nil {
				deps.Logger.Error("LIST_CATALOG_SERVERS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("LIST_CATALOG_SERVERS_SUCCESS", "count", len(items))
			WriteJSON(w, http.StatusOK, map[string]any{"items": items})
		},
	).Methods(http.MethodGet)

	// Add a new catalog server (ADMIN only)
	r.HandleFunc(
		cfg.AdminPrefix+"/catalog/servers",
		func(w http.ResponseWriter, r *http.Request) {
			if ck.GetUserRoleFromContext(r.Context()) != string(m.RoleAdmin) {
				deps.Logger.Error("CREATE_CATALOG_SERVER_FORBIDDEN")
				WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}

			deps.Logger.Info("CREATE_CATALOG_SERVER_INIT")

			var body struct {
				Name        string       `json:"name"`
				URL         string       `json:"url"`
				Description string       `json:"description"`
				AccessType  m.AccessType `json:"access_type"`
				Transport   string       `json:"transport"`
			}
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("CREATE_CATALOG_SERVER_READ_BODY_ERROR")
				return
			}

			if body.Name == "" || body.URL == "" {
				deps.Logger.Error("CREATE_CATALOG_SERVER_MISSING_FIELDS")
				WriteJSON(w, http.StatusBadRequest,
					map[string]string{"error": "missing fields"})
				return
			}

			// Set defaults
			if body.AccessType == "" {
				body.AccessType = m.AccessTypePublic
			}
			if body.Transport == "" {
				body.Transport = "streamable-http"
			}

			rec := m.MCPServer{
				ID:          idgen.NewID(),
				Name:        body.Name,
				URL:         body.URL,
				Description: body.Description,
				AccessType:  body.AccessType,
				Transport:   body.Transport,
			}

			// Use catalog orchestrator to handle auto-fetching for public servers
			serverID, err := deps.CatalogOrchestrator.AddCatalogServer(r.Context(), rec)
			if err != nil {
				deps.Logger.Error("CREATE_CATALOG_SERVER_ORCH_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": err.Error()})
				return
			}

			deps.Logger.Info("CREATE_CATALOG_SERVER_SUCCESS", "id", serverID)
			WriteJSON(w, http.StatusCreated,
				map[string]string{"id": serverID})
		},
	).Methods(http.MethodPost)

	// Update catalog server (ADMIN only)
	r.HandleFunc(
		cfg.AdminPrefix+"/catalog/servers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			if ck.GetUserRoleFromContext(r.Context()) != string(m.RoleAdmin) {
				deps.Logger.Error("UPDATE_CATALOG_SERVER_FORBIDDEN")
				WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			vars := mux.Vars(r)
			id := vars["id"]
			var body struct {
				URL         *string `json:"url"`
				Description *string `json:"description"`
			}
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("UPDATE_CATALOG_SERVER_READ_BODY_ERROR")
				return
			}
			if (body.URL == nil || *body.URL == "") &&
				(body.Description == nil || *body.Description == "") {
				WriteJSON(w, http.StatusBadRequest,
					map[string]string{"error": "no fields to update"})
				return
			}
			url := ""
			desc := ""
			if body.URL != nil {
				url = *body.URL
			}
			if body.Description != nil {
				desc = *body.Description
			}
			if err := deps.Catalog.Update(r.Context(), id, url, desc); err != nil {
				deps.Logger.Error("UPDATE_CATALOG_SERVER_DB_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": err.Error()})
				return
			}

			deps.Logger.Info("UPDATE_CATALOG_SERVER_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		},
	).Methods(http.MethodPatch)

	// Refresh catalog server (ADMIN only) - for public servers only
	r.HandleFunc(
		cfg.AdminPrefix+"/catalog/servers/{id}/refresh",
		func(w http.ResponseWriter, r *http.Request) {
			if ck.GetUserRoleFromContext(r.Context()) != string(m.RoleAdmin) {
				deps.Logger.Error("REFRESH_CATALOG_SERVER_FORBIDDEN")
				WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			vars := mux.Vars(r)
			id := vars["id"]
			deps.Logger.Info("REFRESH_CATALOG_SERVER_INIT", "id", id)

			// Use catalog orchestrator to refresh tools for public servers
			added, deleted, err := deps.CatalogOrchestrator.RefreshCatalogServer(r.Context(), id)
			if err != nil {
				deps.Logger.Error("REFRESH_CATALOG_SERVER_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": err.Error()})
				return
			}

			deps.Logger.Info("REFRESH_CATALOG_SERVER_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]any{
				"ok":            true,
				"added":         added,
				"deleted":       deleted,
				"total_added":   len(added),
				"total_deleted": len(deleted),
			})
		},
	).Methods(http.MethodPost)

	// Get tools for a catalog server (any user for public servers)
	r.HandleFunc(
		cfg.AdminPrefix+"/catalog/servers/{id}/tools",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			serverID := vars["id"]
			deps.Logger.Info("LIST_CATALOG_SERVER_TOOLS_INIT", "server_id", serverID)

			// Get server details to check access_type
			srv, err := deps.Catalog.GetByID(r.Context(), serverID)
			if err != nil {
				deps.Logger.Error("LIST_CATALOG_SERVER_TOOLS_GET_SERVER_ERROR", "error", err)
				WriteJSON(w, http.StatusNotFound,
					map[string]string{"error": "server not found"})
				return
			}

			// Only allow access to public server tools or admin users
			userRole := ck.GetUserRoleFromContext(r.Context())
			if srv.AccessType != m.AccessTypePublic && userRole != string(m.RoleAdmin) {
				deps.Logger.Error("LIST_CATALOG_SERVER_TOOLS_FORBIDDEN")
				WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}

			// Get global tools for this server
			tools, err := deps.Tools.ListGlobalToolsForServer(r.Context(), serverID)
			if err != nil {
				deps.Logger.Error("LIST_CATALOG_SERVER_TOOLS_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": err.Error()})
				return
			}
			deps.Logger.Info("LIST_CATALOG_SERVER_TOOLS_SUCCESS", "count", len(tools))
			WriteJSON(w, http.StatusOK, map[string]any{"items": tools})
		},
	).Methods(http.MethodGet)
}

// Tools routes
func addToolsRoutes(r *mux.Router, deps Deps, cfg Config) {
	// List tools with filters
	r.HandleFunc(
		cfg.AdminPrefix+"/tools",
		func(w http.ResponseWriter, r *http.Request) {
			userID := ck.GetUserIDFromContext(r.Context())
			serverID := r.URL.Query().Get("server_id")        // Primary parameter
			hubServerID := r.URL.Query().Get("hub_server_id") // For filtering by hub
			status := r.URL.Query().Get("status")
			q := r.URL.Query().Get("q")
			deps.Logger.Info("LIST_TOOLS_INIT",
				"user_id", userID,
				"server_id", serverID,
				"hub_server_id", hubServerID,
				"status", status,
				"q_len", len(q),
			)

			if deps.Tools != nil {
				items, err := deps.Tools.ListForUserFiltered(
					r.Context(), userID, serverID, hubServerID, status, q,
				)
				if err != nil {
					deps.Logger.Error("LIST_TOOLS_ERROR", "error", err)
					WriteJSON(
						w,
						http.StatusInternalServerError,
						map[string]string{"error": err.Error()},
					)
					return
				}
				deps.Logger.Info("LIST_TOOLS_SUCCESS", "count", len(items))
				WriteJSON(w, http.StatusOK, map[string]any{"items": items})
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		},
	).Methods(http.MethodGet)

	// Change tool status
	r.HandleFunc(
		cfg.AdminPrefix+"/tools/{id}/status",
		func(w http.ResponseWriter, r *http.Request) {
			type reqBody struct {
				Status string `json:"status"`
			}
			var body reqBody
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("UPDATE_TOOL_STATUS_READ_BODY_ERROR")
				return
			}
			if body.Status == "" {
				deps.Logger.Error("UPDATE_TOOL_STATUS_MISSING_STATUS")
				WriteJSON(
					w,
					http.StatusBadRequest,
					map[string]string{"error": "missing status"},
				)
				return
			}
			id := mux.Vars(r)["id"]
			deps.Logger.Info("UPDATE_TOOL_STATUS_INIT",
				"id", id, "status", body.Status)
			if err := deps.Tools.SetStatus(
				r.Context(), id, body.Status,
			); err != nil {
				deps.Logger.Error("UPDATE_TOOL_STATUS_DB_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("UPDATE_TOOL_STATUS_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodPatch)

	// Soft delete tool
	r.HandleFunc(
		cfg.AdminPrefix+"/tools/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			deps.Logger.Info("DELETE_TOOL_INIT", "id", id)
			if err := deps.Tools.SetStatus(
				r.Context(), id, string(m.StatusDeactivated),
			); err != nil {
				deps.Logger.Error("DELETE_TOOL_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("DELETE_TOOL_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodDelete)
}

// Virtual server routes
func addVirtualServerRoutes(r *mux.Router, deps Deps, cfg Config) {
	// Create with optional tool_ids
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers",
		func(w http.ResponseWriter, r *http.Request) {
			userID := ck.GetUserIDFromContext(r.Context())
			var body struct {
				Name    string   `json:"name"`
				ToolIDs []string `json:"tool_ids"`
			}
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("CREATE_VS_READ_BODY_ERROR")
				return
			}
			deps.Logger.Info("CREATE_VIRTUAL_SERVER_INIT",
				"user_id", userID,
				"tool_ids_len", len(body.ToolIDs))
			var (
				id  string
				err error
			)
			if len(body.ToolIDs) > 0 {
				id, err = deps.Virtual.CreateWithTools(r.Context(), userID, body.Name, body.ToolIDs)
			} else {
				id, err = deps.Virtual.Create(r.Context(), userID, body.Name)
			}
			if err != nil {
				deps.Logger.Error("CREATE_VIRTUAL_SERVER_DB_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("CREATE_VIRTUAL_SERVER_SUCCESS", "id", id)
			WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
		},
	).Methods(http.MethodPost)

	// List for user
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers",
		func(w http.ResponseWriter, r *http.Request) {
			userID := ck.GetUserIDFromContext(r.Context())
			deps.Logger.Info("LIST_VIRTUAL_SERVERS_INIT", "user_id", userID)
			items, err := deps.Virtual.ListForUser(r.Context(), userID)
			if err != nil {
				deps.Logger.Error("LIST_VIRTUAL_SERVERS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("LIST_VIRTUAL_SERVERS_SUCCESS", "count", len(items))
			WriteJSON(w, http.StatusOK, map[string]any{"items": items})
		},
	).Methods(http.MethodGet)

	// Replace tools
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}/tools",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			var body struct {
				ToolIDs []string `json:"tool_ids"`
			}
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("REPLACE_VS_TOOLS_READ_BODY_ERROR")
				return
			}
			deps.Logger.Info("REPLACE_VS_TOOLS_INIT",
				"id", id, "tool_ids_len", len(body.ToolIDs))
			if err := deps.Virtual.ReplaceTools(
				r.Context(), id, body.ToolIDs,
			); err != nil {
				deps.Logger.Error("REPLACE_VS_TOOLS_DB_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("REPLACE_VS_TOOLS_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodPut)

	// Remove one tool from a virtual server
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}/tools/{tool_id}",
		func(w http.ResponseWriter, r *http.Request) {
			vsID := mux.Vars(r)["id"]
			toolID := mux.Vars(r)["tool_id"]
			if vsID == "" || toolID == "" {
				WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing ids"})
				return
			}
			if err := deps.Virtual.RemoveTool(r.Context(), vsID, toolID); err != nil {
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodDelete)

	// List tools for a virtual server
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}/tools",
		func(w http.ResponseWriter, r *http.Request) {
			vsID := mux.Vars(r)["id"]
			deps.Logger.Info("LIST_VS_TOOLS_INIT", "id", vsID)
			items, err := deps.Tools.ListForVirtualServer(
				r.Context(), vsID,
			)
			if err != nil {
				deps.Logger.Error("LIST_VS_TOOLS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("LIST_VS_TOOLS_SUCCESS", "count", len(items))
			WriteJSON(w, http.StatusOK, map[string]any{"items": items})
		},
	).Methods(http.MethodGet)

	// Set status
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}/status",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			var body struct {
				Status string `json:"status"`
			}
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("UPDATE_VS_STATUS_READ_BODY_ERROR")
				return
			}
			if body.Status == "" {
				deps.Logger.Error("UPDATE_VS_STATUS_MISSING_STATUS")
				WriteJSON(
					w,
					http.StatusBadRequest,
					map[string]string{"error": "missing status"},
				)
				return
			}

			deps.Logger.Info("UPDATE_VS_STATUS_INIT",
				"id", id, "status", body.Status)
			if err := deps.Virtual.SetStatus(
				r.Context(), id, body.Status,
			); err != nil {
				deps.Logger.Error("UPDATE_VS_STATUS_DB_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}
			deps.Logger.Info("UPDATE_VS_STATUS_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodPatch)

	// Update virtual server properties (name)
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			var body struct {
				Name *string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
				return
			}
			deps.Logger.Info("UPDATE_VS_INIT", "id", id, "name", body.Name)

			if body.Name != nil && *body.Name == "" {
				WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Name cannot be empty"})
				return
			}

			if err := deps.Virtual.UpdateName(r.Context(), id, body.Name); err != nil {
				deps.Logger.Error("UPDATE_VS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("UPDATE_VS_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "updated"})
		},
	).Methods(http.MethodPatch)

	// Delete
	r.HandleFunc(
		cfg.AdminPrefix+"/virtual-servers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			deps.Logger.Info("DELETE_VS_INIT", "id", id)
			if err := deps.Virtual.Delete(r.Context(), id); err != nil {
				deps.Logger.Error("DELETE_VS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("DELETE_VS_SUCCESS", "id", id)
			WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
		},
	).Methods(http.MethodDelete)
}

// Hub routes
func addHubRoutes(r *mux.Router, deps Deps, cfg Config) {
	orch := deps.McphubOrchestrator
	// List user's hubs
	r.HandleFunc(
		cfg.AdminPrefix+"/hub/servers",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("LIST_HUB_SERVERS_INIT",
				"user_id", ck.GetUserIDFromContext(r.Context()))
			userID := ck.GetUserIDFromContext(r.Context())
			items, err := deps.Hubs.ListForUser(r.Context(), userID)
			if err != nil {
				deps.Logger.Error("LIST_HUB_SERVERS_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("LIST_HUB_SERVERS_SUCCESS", "count", len(items))
			WriteJSON(w, http.StatusOK, map[string]any{"items": items})
		},
	).Methods(http.MethodGet)

	// Add hub server via orchestrator
	r.HandleFunc(
		cfg.AdminPrefix+"/hub/servers",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("CREATE_HUB_SERVER_INIT")
			var body orchestrator.CreateMCPHubServer
			if !ReadJSON(w, r, &body) {
				deps.Logger.Error("CREATE_HUB_SERVER_READ_BODY_ERROR")
				return
			}

			deps.Logger.Info("CREATE_HUB_SERVER_REQ", "req", body)

			// Always trust server-side authenticated user
			body.UserID = ck.GetUserIDFromContext(r.Context())

			id, err := orch.AddHub(r.Context(), body)
			if err != nil {
				deps.Logger.Error("CREATE_HUB_SERVER_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("CREATE_HUB_SERVER_SUCCESS", "id", id)
			WriteJSON(w, http.StatusCreated, map[string]any{"id": id, "ok": true})
		},
	).Methods(http.MethodPost)

	// Update status / Delete hub
	r.HandleFunc(
		cfg.AdminPrefix+"/hub/servers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			switch r.Method {
			case http.MethodDelete:
				deps.Logger.Info("DELETE_HUB_SERVER_INIT", "id", id)
				if err := deps.Hubs.Delete(r.Context(), id); err != nil {
					deps.Logger.Error("DELETE_HUB_SERVER_ERROR", "error", err)
					WriteJSON(
						w,
						http.StatusInternalServerError,
						map[string]string{"error": err.Error()},
					)
					return
				}

				deps.Logger.Info("DELETE_HUB_SERVER_SUCCESS", "id", id)
				WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
			case http.MethodPatch:
				type req struct {
					Status string `json:"status"`
				}
				var b req
				if !ReadJSON(w, r, &b) {
					deps.Logger.Error("UPDATE_HUB_STATUS_READ_BODY_ERROR")
					return
				}
				if b.Status == "" {
					deps.Logger.Error("UPDATE_HUB_STATUS_MISSING_STATUS")
					WriteJSON(
						w,
						http.StatusBadRequest,
						map[string]string{"error": "missing status"},
					)
					return
				}

				deps.Logger.Info("UPDATE_HUB_STATUS_INIT",
					"id", id, "status", b.Status)
				if err := deps.Hubs.SetStatus(
					r.Context(), id, b.Status,
				); err != nil {
					deps.Logger.Error("UPDATE_HUB_STATUS_DB_ERROR", "error", err)
					WriteJSON(
						w,
						http.StatusInternalServerError,
						map[string]string{"error": err.Error()},
					)
					return
				}
				deps.Logger.Info("UPDATE_HUB_STATUS_SUCCESS", "id", id)
				WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		},
	).Methods(http.MethodDelete, http.MethodPatch)

	// Refresh tools for a hub via orchestrator
	r.HandleFunc(
		cfg.AdminPrefix+"/hub/servers/{id}/refresh",
		func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			userID := ck.GetUserIDFromContext(r.Context())
			deps.Logger.Info("REFRESH_HUB_INIT", "id", id, "user_id", userID)

			added, deleted, err := orch.RefreshHub(r.Context(), id, userID)
			if err != nil {
				deps.Logger.Error("REFRESH_HUB_ERROR", "error", err)
				WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": err.Error()},
				)
				return
			}

			deps.Logger.Info("REFRESH_HUB_SUCCESS",
				"added", len(added),
				"deleted", len(deleted),
			)

			WriteJSON(w, http.StatusOK, map[string]any{
				"ok":            true,
				"added":         added,
				"deleted":       deleted,
				"total_added":   len(added),
				"total_deleted": len(deleted),
			})
		},
	).Methods(http.MethodPost)
}
