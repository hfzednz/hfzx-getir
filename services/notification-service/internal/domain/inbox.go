package domain

import (
	"time"

	"github.com/google/uuid"
)

// InboxItem is an in-app notification.
type InboxItem struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	MessageID   uuid.UUID
	Title       string
	Body        string
	Read        bool
	Archived    bool
	CreatedAt   time.Time
	ReadAt      *time.Time
}
