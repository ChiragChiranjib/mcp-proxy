// Package mcphub_orchestrator coordinates higher-level workflows around MCP hubs.
package mcphub_orchestrator

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	mcpclient "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/client"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Orchestrator wires the client with services and DB for
// transactional workflows.
type Orchestrator struct {
	hubs   *mcphub.Service
	tools  *tool.Service
	repo   *repo.Repo
	logger *slog.Logger
	encr   *encryptor.AESEncrypter
}

// CreateMCPHubServer captures inputs needed to create a hub.
type CreateMCPHubServer struct {
	UserID      string          `json:"user_id,omitempty"`
	MCPServerID string          `json:"mcp_server_id"`
	AuthType    m.AuthType      `json:"auth_type"`
	AuthValue   json.RawMessage `json:"auth_value"`
}

// New creates an orchestrator.
func New(
	hubs *mcphub.Service,
	tools *tool.Service,
	r *repo.Repo,
	logger *slog.Logger,
	encr *encryptor.AESEncrypter,
) *Orchestrator {
	return &Orchestrator{
		hubs:   hubs,
		tools:  tools,
		repo:   r,
		logger: logger,
		encr:   encr,
	}
}

// AddHub creates a hub and its tools atomically
// after discovering capabilities and tools.
func (o *Orchestrator) AddHub(
	ctx context.Context, req CreateMCPHubServer) (string, error) {
	o.logger.Info("ORCH_ADD_HUB_INIT",
		"user_id", req.UserID,
		"mcp_server_id", req.MCPServerID,
		"auth_type", req.AuthType,
	)
	// Resolve MCP server URL and name from catalog via repo
	srv, err := o.repo.GetCatalogServerByID(ctx, req.MCPServerID)
	if err != nil {
		o.logger.Error("ORCH_RESOLVE_SERVER_ERROR", "error", err)
		return "", err
	}
	serverURL := srv.URL
	serverName := srv.Name
	o.logger.Info("ORCH_RESOLVE_SERVER_SUCCESS",
		"server_name", serverName, "server_url_len", len(serverURL))

	// Build hub model
	hubID := idgen.NewID()
	hub := m.MCPHubServer{
		ID:          hubID,
		UserID:      req.UserID,
		MCPServerID: req.MCPServerID,
		Status:      m.StatusActive,
		AuthType:    req.AuthType,
		AuthValue:   req.AuthValue,
	}

	// Encrypt auth value if provided (both bearer tokens and custom headers)
	if (req.AuthType == m.AuthTypeBearer || req.AuthType == m.AuthTypeCustomHeaders) &&
		len(req.AuthValue) > 0 && o.encr != nil {
		o.logger.Info("ORCH_ENCRYPT_AUTH_INIT", "auth_type", req.AuthType)
		enc, err := o.encr.EncryptToJSON(req.AuthValue)
		if err != nil {
			o.logger.Error("ORCH_ENCRYPT_AUTH_ERROR", "error", err, "auth_type", req.AuthType)
			return "", err
		}
		hub.AuthValue = enc
		o.logger.Info("ORCH_ENCRYPT_AUTH_SUCCESS", "len", len(enc), "auth_type", req.AuthType)
	}

	// For public servers, skip tool fetching since global tools already exist
	// For private servers, fetch capabilities and tools with user-specific auth
	var toolModels []m.MCPTool
	if srv.AccessType == m.AccessTypePrivate {
		headers := mcpclient.BuildUpstreamHeaders(o.logger, o.encr, &hub)

		// Fetch capabilities via init and tools via client
		o.logger.Info("ORCH_INIT_CAPABILITIES_INIT", "server_url", serverURL, "access_type", srv.AccessType)
		caps, err := mcpclient.InitCapabilities(ctx, serverURL, headers)
		if err != nil {
			o.logger.Error("ORCH_INIT_CAPABILITIES_ERROR", "error", err)
			return "", err
		}
		o.logger.Info("ORCH_INIT_CAPABILITIES_SUCCESS", "len", len(caps))

		o.logger.Info("ORCH_LIST_TOOLS_INIT")
		toolsRes, err := mcpclient.ListTools(ctx, serverURL, headers)
		if err != nil {
			o.logger.Error("ORCH_LIST_TOOLS_ERROR", "error", err)
			return "", err
		}
		o.logger.Info("ORCH_LIST_TOOLS_SUCCESS", "tool_count", len(toolsRes.Tools))

		// Build tool models (user-specific tools for private servers)
		for _, t := range toolsRes.Tools {
			mod := serverName + "-" + t.Name
			schemaJSON, _ := json.Marshal(t.InputSchema)
			annotationsJSON, _ := json.Marshal(t.Annotations)
			toolModels = append(toolModels, m.MCPTool{
				ID:             idgen.NewID(),
				UserID:         &req.UserID, // User-specific tool
				MCPServerID:    req.MCPServerID,
				MCPHubServerID: &hubID, // Link to the hub server
				OriginalName:   t.Name,
				ModifiedName:   mod,
				Description:    t.Description,
				InputSchema:    schemaJSON,
				Annotations:    annotationsJSON,
				Status:         m.StatusActive,
			})
		}
	} else {
		o.logger.Info("ORCH_SKIP_TOOL_FETCH", "access_type", srv.AccessType, "reason", "global tools already exist")
	}
	o.logger.Info("ORCH_BUILD_TOOL_MODELS_SUCCESS", "count", len(toolModels))

	// Transaction: create hub and its tools
	o.logger.Info("ORCH_ADD_HUB_TX_BEGIN")
	err = o.repo.Transaction(func(tx *repo.Repo) error {
		if err := tx.WithContext(ctx).Create(&hub).Error; err != nil {
			return err
		}
		if len(toolModels) > 0 {
			if err := tx.WithContext(ctx).Create(&toolModels).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		o.logger.Error("ORCH_ADD_HUB_TX_ERROR", "error", err)
		return "", err
	}
	o.logger.Info("ORCH_ADD_HUB_SUCCESS",
		"hub_id", hubID, "tool_count", len(toolModels))
	return hubID, nil
}

// RefreshHub reconciles tools for a hub against upstream and returns details
// of tools added and deleted.
func (o *Orchestrator) RefreshHub(
	ctx context.Context,
	hubID string,
	userID string,
) (added []m.MCPTool, deleted []m.MCPTool, err error) {
	o.logger.Info("ORCH_REFRESH_INIT", "hub_id", hubID, "user_id", userID)
	info, err := o.hubs.GetWithURL(ctx, hubID)
	if err != nil {
		o.logger.Error("ORCH_REFRESH_GET_WITH_URL_ERROR", "error", err)
		return nil, nil, err
	}

	serverURL := info.URL
	serverName := info.Name

	// For public servers, tools are managed globally, not per hub
	if info.AccessType == m.AccessTypePublic {
		o.logger.Info("ORCH_REFRESH_SKIP_PUBLIC", "access_type", info.AccessType, "reason", "public servers managed globally")
		return nil, nil, nil // No tools to add/delete for public servers
	}

	o.logger.Info("ORCH_REFRESH_LIST_TOOLS_INIT", "access_type", info.AccessType)
	headers := mcpclient.BuildUpstreamHeaders(o.logger, o.encr, &info.MCPHubServer)
	res, err := mcpclient.ListTools(ctx, serverURL, headers)
	if err != nil {
		o.logger.Error("ORCH_REFRESH_LIST_TOOLS_ERROR", "error", err)
		return nil, nil, err
	}
	o.logger.Info("ORCH_REFRESH_LIST_TOOLS_SUCCESS", "tool_count", len(res.Tools))

	// Desired set
	desired := make(map[string]mcp.Tool)
	for _, t := range res.Tools {
		desired[serverName+"-"+t.Name] = t
	}

	// Current set from DB (user-specific tools for this server)
	var current []m.MCPTool
	o.logger.Info("ORCH_REFRESH_DB_LOAD_TOOLS_INIT")
	if err := o.repo.WithContext(ctx).
		Where("mcp_server_id = ? AND user_id = ?", info.MCPServerID, userID).
		Find(&current).Error; err != nil {
		o.logger.Error("ORCH_REFRESH_DB_LOAD_TOOLS_ERROR", "error", err)
		return nil, nil, err
	}
	o.logger.Info("ORCH_REFRESH_DB_LOAD_TOOLS_SUCCESS",
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
				o.logger.Error("ORCH_REFRESH_ANNOTATIONS_MARSHALL_ERROR",
					"error", err)
			}

			toInsert = append(toInsert, m.MCPTool{
				ID:             idgen.NewID(),
				UserID:         &userID, // User-specific tool
				MCPServerID:    info.MCPServerID,
				MCPHubServerID: &hubID, // Link to the hub server
				OriginalName:   dtool.Name,
				ModifiedName:   serverName + "-" + dtool.Name,
				Description:    dtool.Description,
				InputSchema:    dtool.RawInputSchema,
				Annotations:    annotationsJSON,
				Status:         m.StatusActive,
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
	o.logger.Info("ORCH_REFRESH_TX_BEGIN",
		"to_add", len(toInsert), "to_delete", len(toDeleteIDs))
	err = o.repo.Transaction(func(tx *repo.Repo) error {
		if len(toInsert) > 0 {
			if err := tx.WithContext(ctx).Create(&toInsert).Error; err != nil {
				return err
			}
		}
		if len(toDeleteIDs) > 0 {
			if err := tx.WithContext(ctx).
				Where("id IN ?", toDeleteIDs).
				Delete(&m.MCPTool{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		o.logger.Error("ORCH_REFRESH_TX_ERROR", "error", err)
		return nil, nil, err
	}
	// return what was added and deleted
	o.logger.Info("ORCH_REFRESH_SUCCESS",
		"added", len(toInsert), "deleted", len(deleted))
	return toInsert, deleted, nil
}
