// Package idgen provides helpers for generating public IDs.
package idgen

import (
	"crypto/rand"
	"encoding/base64"
)

// NewID returns a URL-safe, 22-character base64 identifier.
func NewID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}
