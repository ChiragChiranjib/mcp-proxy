package client

import (
	"encoding/json"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// BuildUpstreamHeaders prepares Authorization/custom headers,
// with AES decryption when bearer token is stored encrypted as JSON.
func BuildUpstreamHeaders(
	logger *slog.Logger,
	encypt *encryptor.AESEncrypter,
	hub *m.MCPHubServer,
) map[string]string {
	logger.Info(
		"BUILD_UPSTREAM_HEADERS_INIT",
		"hub_id", hub.ID,
		"auth_type", string(hub.AuthType),
	)
	headers := map[string]string{}
	switch hub.AuthType {
	case m.AuthTypeBearer:
		token := string(hub.AuthValue)
		if len(hub.AuthValue) > 0 && hub.AuthValue[0] == '{' {
			if b, err := encypt.DecryptFromJSON(hub.AuthValue); err == nil {
				token = string(b)
				logger.Info("DECRYPT_BEARER_TOKEN_OK", "len", len(token))
			} else {
				logger.Error("DECRYPT_BEARER_TOKEN_ERROR", "error", err)
			}
		}
		headers["Authorization"] = "Bearer " + token
	case m.AuthTypeCustomHeaders:
		var hdrs map[string]string
		if err := json.Unmarshal(hub.AuthValue, &hdrs); err != nil {
			logger.Error("CUSTOM_HEADERS_DECODE_ERROR", "error", err)
		} else {
			for k, v := range hdrs {
				headers[k] = v
			}
			logger.Info("CUSTOM_HEADERS_APPLIED", "count", len(headers))
		}
	default:
		logger.Info("NO_AUTH_HEADERS_APPLIED")
	}
	return headers
}