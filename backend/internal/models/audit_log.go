package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        int64
	TenantID  uuid.UUID
	UserID    *uuid.UUID // nil for guest actions
	Action    string
	Diff      json.RawMessage
	CreatedAt time.Time
}
