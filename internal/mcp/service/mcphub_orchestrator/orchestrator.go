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
	Transport   string          `json:"transport"`
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
		"transport", req.Transport,
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
		ID:           hubID,
		UserID:       req.UserID,
		MCPServerID:  req.MCPServerID,
		Status:       m.StatusActive,
		Transport:    req.Transport,
		Capabilities: nil,
		AuthType:     req.AuthType,
		AuthValue:    req.AuthValue,
	}

	headers := mcpclient.BuildUpstreamHeaders(o.logger, o.encr, &hub)

	// Fetch capabilities via init and tools via client
	o.logger.Info("ORCH_INIT_CAPABILITIES_INIT", "server_url", serverURL)
	caps, err := mcpclient.InitCapabilities(ctx, serverURL, headers)
	if err != nil {
		o.logger.Error("ORCH_INIT_CAPABILITIES_ERROR", "error", err)
		return "", err
	}
	hub.Capabilities = caps
	o.logger.Info("ORCH_INIT_CAPABILITIES_SUCCESS", "len", len(caps))

	o.logger.Info("ORCH_LIST_TOOLS_INIT")
	toolsRes, err := mcpclient.ListTools(ctx, serverURL, headers)
	if err != nil {
		o.logger.Error("ORCH_LIST_TOOLS_ERROR", "error", err)
		return "", err
	}
	o.logger.Info("ORCH_LIST_TOOLS_SUCCESS", "tool_count", len(toolsRes.Tools))

	// Encrypt bearer token if provided
	if req.AuthType == m.AuthTypeBearer &&
		len(req.AuthValue) > 0 && o.encr != nil {
		o.logger.Info("ORCH_ENCRYPT_BEARER_INIT")
		enc, err := o.encr.EncryptToJSON(req.AuthValue)
		if err != nil {
			o.logger.Error("ORCH_ENCRYPT_BEARER_ERROR", "error", err)
			return "", err
		}
		hub.AuthValue = enc
		o.logger.Info("ORCH_ENCRYPT_BEARER_SUCCESS", "len", len(enc))
	}

	// Build tool models
	var toolModels []m.MCPTool
	for _, t := range toolsRes.Tools {
		mod := serverName + "-" + t.Name
		schemaJSON, _ := json.Marshal(t.InputSchema)
		annotationsJSON, _ := json.Marshal(t.Annotations)
		toolModels = append(toolModels, m.MCPTool{
			ID:             idgen.NewID(),
			UserID:         req.UserID,
			OriginalName:   t.Name,
			ModifiedName:   mod,
			MCPHubServerID: hubID,
			Description:    t.Description,
			InputSchema:    schemaJSON,
			Annotations:    annotationsJSON,
			Status:         m.StatusActive,
		})
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
	o.logger.Info("ORCH_REFRESH_LIST_TOOLS_INIT")
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

	// Current set from DB
	var current []m.MCPTool
	o.logger.Info("ORCH_REFRESH_DB_LOAD_TOOLS_INIT")
	if err := o.repo.WithContext(ctx).
		Where("mcp_hub_server_id = ?", hubID).
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
				UserID:         userID,
				OriginalName:   dtool.Name,
				ModifiedName:   serverName + "-" + dtool.Name,
				MCPHubServerID: hubID,
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
