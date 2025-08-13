// Package catalog_orchestrator coordinates higher-level workflows around catalog servers.
package catalog_orchestrator

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	mcpclient "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/client"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
)

// Orchestrator wires the catalog service with tools service for
// transactional workflows.
type Orchestrator struct {
	catalog *catalog.Service
	tools   *tool.Service
	repo    *repo.Repo
	logger  *slog.Logger
	encr    *encryptor.AESEncrypter
}

// New creates a catalog orchestrator.
func New(
	catalogSvc *catalog.Service,
	toolsSvc *tool.Service,
	r *repo.Repo,
	logger *slog.Logger,
	encr *encryptor.AESEncrypter,
) *Orchestrator {
	return &Orchestrator{
		catalog: catalogSvc,
		tools:   toolsSvc,
		repo:    r,
		logger:  logger,
		encr:    encr,
	}
}

// AddCatalogServer creates a catalog server and optionally fetches tools if access_type is public.
func (o *Orchestrator) AddCatalogServer(
	ctx context.Context, srv m.MCPServer) (string, error) {
	o.logger.Info("CATALOG_ORCH_ADD_SERVER_INIT",
		"name", srv.Name,
		"access_type", srv.AccessType,
	)

	var caps []byte
	var toolModels []m.MCPTool

	// If access_type is public, fetch capabilities and tools
	if srv.AccessType == m.AccessTypePublic {
		serverURL := srv.URL
		serverName := srv.Name

		// No auth headers for public servers
		headers := map[string]string{}

		// Fetch capabilities
		o.logger.Info("CATALOG_ORCH_INIT_CAPABILITIES_INIT", "server_url", serverURL)
		var err error
		caps, err = mcpclient.InitCapabilities(ctx, serverURL, headers)
		if err != nil {
			o.logger.Error("CATALOG_ORCH_INIT_CAPABILITIES_ERROR", "error", err)
			return "", err
		}
		o.logger.Info("CATALOG_ORCH_INIT_CAPABILITIES_SUCCESS", "len", len(caps))

		srv.Capabilities = caps

		// Fetch tools
		o.logger.Info("CATALOG_ORCH_LIST_TOOLS_INIT")
		toolsRes, err := mcpclient.ListTools(ctx, serverURL, headers)
		if err != nil {
			o.logger.Error("CATALOG_ORCH_LIST_TOOLS_ERROR", "error", err)
			return "", err
		}
		o.logger.Info("CATALOG_ORCH_LIST_TOOLS_SUCCESS", "tool_count", len(toolsRes.Tools))

		// Build tool models (global tools with user_id = nil)
		for _, t := range toolsRes.Tools {
			mod := serverName + "-" + t.Name
			schemaJSON, _ := json.Marshal(t.InputSchema)
			annotationsJSON, _ := json.Marshal(t.Annotations)
			toolModels = append(toolModels, m.MCPTool{
				ID:           idgen.NewID(),
				UserID:       nil, // Global tool
				MCPServerID:  srv.ID,
				OriginalName: t.Name,
				ModifiedName: mod,
				Description:  t.Description,
				InputSchema:  schemaJSON,
				Annotations:  annotationsJSON,
				Status:       m.StatusActive,
			})
		}
		o.logger.Info("CATALOG_ORCH_BUILD_TOOL_MODELS_SUCCESS", "count", len(toolModels))
	}

	// Transaction: create server and create tools
	o.logger.Info("CATALOG_ORCH_ADD_SERVER_TX_BEGIN")
	err := o.repo.Transaction(func(tx *repo.Repo) error {
		// Create the server
		if err := tx.CreateCatalogServer(ctx, srv); err != nil {
			return err
		}

		// Create tools if we have them
		if err := tx.CreateTools(ctx, toolModels); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		o.logger.Error("CATALOG_ORCH_ADD_SERVER_TX_ERROR", "error", err)
		return "", err
	}

	o.logger.Info("CATALOG_ORCH_ADD_SERVER_SUCCESS",
		"server_id", srv.ID, "tool_count", len(toolModels))
	return srv.ID, nil
}

// RefreshCatalogServer refreshes tools for a public catalog server and returns details
// of tools added and deleted.
func (o *Orchestrator) RefreshCatalogServer(
	ctx context.Context, serverID string) (added []m.MCPTool, deleted []m.MCPTool, err error) {
	o.logger.Info("CATALOG_ORCH_REFRESH_INIT", "server_id", serverID)

	// Get server details
	srv, err := o.catalog.GetByID(ctx, serverID)
	if err != nil {
		o.logger.Error("CATALOG_ORCH_REFRESH_GET_SERVER_ERROR", "error", err)
		return nil, nil, err
	}

	// Only allow refresh for public servers
	if srv.AccessType != m.AccessTypePublic {
		o.logger.Error("CATALOG_ORCH_REFRESH_NOT_PUBLIC", "access_type", srv.AccessType)
		return nil, nil, nil // Not an error, just skip
	}

	o.logger.Info("CATALOG_ORCH_REFRESH_FETCH_TOOLS_INIT")
	added, deleted, err = o.fetchAndStoreToolsWithDiff(ctx, srv)
	if err != nil {
		o.logger.Error("CATALOG_ORCH_REFRESH_FETCH_TOOLS_ERROR", "error", err)
		return nil, nil, err
	}

	o.logger.Info("CATALOG_ORCH_REFRESH_SUCCESS",
		"server_id", serverID, "added", len(added), "deleted", len(deleted))
	return added, deleted, nil
}

// fetchAndStoreToolsWithDiff connects to a public server, fetches tools,
// and returns what was added/deleted.
func (o *Orchestrator) fetchAndStoreToolsWithDiff(
	ctx context.Context, srv m.MCPServer) (added []m.MCPTool, deleted []m.MCPTool, err error) {
	serverURL := srv.URL
	serverName := srv.Name

	// No auth headers for public servers
	headers := map[string]string{}

	// Fetch capabilities
	o.logger.Info("CATALOG_ORCH_INIT_CAPABILITIES_INIT", "server_url", serverURL)
	caps, err := mcpclient.InitCapabilities(ctx, serverURL, headers)
	if err != nil {
		o.logger.Error("CATALOG_ORCH_INIT_CAPABILITIES_ERROR", "error", err)
		return nil, nil, err
	}
	o.logger.Info("CATALOG_ORCH_INIT_CAPABILITIES_SUCCESS", "len", len(caps))

	// Update server with capabilities
	if err := o.catalog.UpdateCapabilities(ctx, srv.ID, caps, srv.Transport); err != nil {
		o.logger.Error("CATALOG_ORCH_UPDATE_CAPABILITIES_ERROR", "error", err)
		return nil, nil, err
	}

	// Fetch tools
	o.logger.Info("CATALOG_ORCH_LIST_TOOLS_INIT")
	toolsRes, err := mcpclient.ListTools(ctx, serverURL, headers)
	if err != nil {
		o.logger.Error("CATALOG_ORCH_LIST_TOOLS_ERROR", "error", err)
		return nil, nil, err
	}
	o.logger.Info("CATALOG_ORCH_LIST_TOOLS_SUCCESS", "tool_count", len(toolsRes.Tools))

	// Desired set
	desired := make(map[string]mcp.Tool)
	for _, t := range toolsRes.Tools {
		desired[serverName+"-"+t.Name] = t
	}

	// Current set from DB (global tools for this server)
	o.logger.Info("CATALOG_ORCH_REFRESH_DB_LOAD_TOOLS_INIT")
	current, err := o.repo.ListGlobalToolsForServer(ctx, srv.ID)
	if err != nil {
		o.logger.Error("CATALOG_ORCH_REFRESH_DB_LOAD_TOOLS_ERROR",
			"error", err)
		return nil, nil, err
	}
	o.logger.Info("CATALOG_ORCH_REFRESH_DB_LOAD_TOOLS_SUCCESS",
		"current_count", len(current))

	currentSet := make(map[string]m.MCPTool)
	for _, t := range current {
		currentSet[t.ModifiedName] = t
	}

	// Compute to-add and to-remove
	var toInsert []m.MCPTool
	for name, dtool := range desired {
		if _, ok := currentSet[name]; !ok {
			annotationsJSON, err := json.Marshal(dtool.Annotations)
			if err != nil {
				o.logger.Error("CATALOG_ORCH_REFRESH_ANNOTATIONS_MARSHALL_ERROR",
					"error", err)
			}

			toInsert = append(toInsert, m.MCPTool{
				ID:           idgen.NewID(),
				UserID:       nil, // Global tool
				MCPServerID:  srv.ID,
				OriginalName: dtool.Name,
				ModifiedName: serverName + "-" + dtool.Name,
				Description:  dtool.Description,
				InputSchema:  dtool.RawInputSchema,
				Annotations:  annotationsJSON,
				Status:       m.StatusActive,
			})
		}
	}

	var toDeleteIDs []string
	for name, rec := range currentSet {
		if _, ok := desired[name]; !ok {
			toDeleteIDs = append(toDeleteIDs, rec.ID)
			deleted = append(deleted, rec)
		}
	}

	// Apply changes transactionally
	o.logger.Info("CATALOG_ORCH_REFRESH_TX_BEGIN",
		"to_add", len(toInsert), "to_delete", len(toDeleteIDs))
	err = o.repo.Transaction(func(tx *repo.Repo) error {
		if err := tx.CreateTools(ctx, toInsert); err != nil {
			return err
		}
		if err := tx.DeleteToolsByIDs(ctx, toDeleteIDs); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		o.logger.Error("CATALOG_ORCH_REFRESH_TX_ERROR",
			"error", err)
		return nil, nil, err
	}

	// Return what was added and deleted
	o.logger.Info("CATALOG_ORCH_REFRESH_TOOLS_SUCCESS",
		"added", len(toInsert), "deleted", len(deleted))
	return toInsert, deleted, nil
}
