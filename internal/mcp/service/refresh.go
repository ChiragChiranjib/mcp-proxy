// Package service contains application-level orchestrators across MCP subdomains.
package service

import (
	"context"
	"encoding/json"
	"strconv"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

// RefreshHubTools connects to the upstream MCP server for the given hub server,
// lists tools, and upserts them for the given user. It returns the number of tools upserted.
func RefreshHubTools(ctx context.Context, hubs *mcphub.Service, tools *tool.Service, hubID string, userID string) (int, error) {
	// Resolve upstream URL and auth
	info, err := hubs.GetWithURL(ctx, hubID)
	if err != nil {
		return 0, err
	}

	// Build client transport (auth headers handled in proxy path later if needed)
	transport := sdk.NewStreamableClientTransport(info.ServerURL, nil)
	client := sdk.NewClient(&sdk.Implementation{Name: "mcp-proxy-refresh", Version: "0.1.0"}, nil)
	cs, err := client.Connect(ctx, transport)
	if err != nil {
		return 0, err
	}
	defer func() { _ = cs.Close() }()

	// List tools from upstream
	res, err := cs.ListTools(ctx, &sdk.ListToolsParams{})
	if err != nil {
		return 0, err
	}

	count := 0
	// Build a set of existing modified names for dedupe
	existing, _ := tools.ListActiveForHub(ctx, hubID)
	used := make(map[string]struct{}, len(existing))
	for _, t := range existing {
		used[t.ModifiedName] = struct{}{}
	}

	for _, t := range res.Tools {
		// Dedupe naming strategy: use original name and suffix when necessary
		base := t.Name
		modified := base
		i := 2
		for {
			if _, ok := used[modified]; !ok {
				break
			}
			modified = base + "__" + strconv.Itoa(i)
			i++
		}
		used[modified] = struct{}{}
		// Upsert using original name and derived modified name
		schemaJSON, _ := json.Marshal(t.InputSchema)
		if err := tools.Upsert(ctx, types.Tool{
			ID:           "", // let DB layer handle new IDs or provide external generation upstream
			UserID:       userID,
			OriginalName: t.Name,
			ModifiedName: modified,
			HubServerID:  hubID,
			InputSchema:  schemaJSON,
			Annotations:  nil,
			Status:       "ACTIVE",
		}); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
