package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        int64           `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	Action    string          `json:"action"`
	Diff      json.RawMessage `json:"diff,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
