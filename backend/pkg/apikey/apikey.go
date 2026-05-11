package apikey

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

const version = "v1"

// Generate creates a signed API key that embeds the project ID.
// Format: v1.<base64url(project_id_bytes)>.<base64url(hmac-sha256)>
func Generate(projectID uuid.UUID, secret string) string {
	payload := base64.RawURLEncoding.EncodeToString(projectID[:])
	sig := sign(payload, secret)
	return version + "." + payload + "." + sig
}

// Parse validates the API key signature and returns the embedded project ID.
// No database lookup required.
func Parse(key, secret string) (uuid.UUID, bool) {
	parts := strings.SplitN(key, ".", 3)
	if len(parts) != 3 || parts[0] != version {
		return uuid.UUID{}, false
	}

	payload, sig := parts[1], parts[2]

	expected := sign(payload, secret)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return uuid.UUID{}, false
	}

	b, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil || len(b) != 16 {
		return uuid.UUID{}, false
	}

	var id uuid.UUID
	copy(id[:], b)
	return id, true
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte("apikey:"+secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
