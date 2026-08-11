package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// MarkInboxRead marks an inbox item as read.
func (d *Deps) MarkInboxRead(ctx context.Context, tenantID, itemID uuid.UUID) (domain.InboxItem, error) {
	if tenantID == uuid.Nil || itemID == uuid.Nil {
		return domain.InboxItem{}, fmt.Errorf("%w: tenant_id and item_id required", domain.ErrInvalidArgument)
	}
	return d.Inbox.MarkRead(ctx, tenantID, itemID, d.now())
}

// ListInbox lists inbox items for a principal.
func (d *Deps) ListInbox(ctx context.Context, tenantID, principalID uuid.UUID, includeArchived bool) ([]domain.InboxItem, error) {
	if tenantID == uuid.Nil || principalID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	return d.Inbox.List(ctx, tenantID, principalID, includeArchived)
}
