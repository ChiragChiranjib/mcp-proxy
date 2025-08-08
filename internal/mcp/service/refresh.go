// Package service contains application-level orchestrators across MCP subdomains.
package service

import (
	"context"
	"encoding/json"
	"strconv"

	mcpclient "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/client"
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

	// List tools from upstream via client helper
	res, err := mcpclient.ListTools(ctx, info.ServerURL)
	if err != nil {
		return 0, err
	}

	count := 0
	// Build a set of existing modified names for dedupe
	// We no longer need to prefetch existing tools for naming; we insert blindly.
	used := make(map[string]struct{})

	for _, t := range res.Tools {
		// Preferred naming: <server-name>-<tool-name> (normalized) and if duplicate in this run, suffix
		base := info.ServerName + "-" + t.Name
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
