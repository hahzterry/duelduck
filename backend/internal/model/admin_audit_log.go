package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
)

type AdminAuditLog struct {
	ch.CHModel `ch:"table:admin_audit_logs"`

	ModeratorID uuid.UUID `ch:"moderator_id"`
	Role        uint8     `ch:"role"`
	Action      string    `ch:"action"`
	IP          string    `ch:"ip"`
	Metadata    string    `ch:"metadata"`
	CreatedAt   time.Time `ch:"created_at"`
}

const (
	AuditActionGetUser           = "get_user"
	AuditActionBlockAPIKey       = "block_api_key"
	AuditActionUnblockAPIKey     = "unblock_api_key"
	AuditActionCreateProject     = "create_project"
	AuditActionCreateCryptoDuel  = "create_crypto_duel"
	AuditActionApproveCryptoDuel = "approve_crypto_duel"
	AuditActionResolveCryptoDuel = "resolve_crypto_duel"
	AuditActionCancelCryptoDuel  = "cancel_crypto_duel"
	AuditActionEditDuel          = "edit_duel"
)

func AuditMeta(kv ...any) string {
	if len(kv)%2 != 0 {
		kv = kv[:len(kv)-1]
	}
	m := make(map[string]any, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		if key, ok := kv[i].(string); ok {
			m[key] = kv[i+1]
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}
