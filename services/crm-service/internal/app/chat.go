package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/domain"
)

// StartChatInput starts a live chat conversation.
type StartChatInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Channel    string
	TicketID   *uuid.UUID
	AgentID    *uuid.UUID
}

// StartChat creates an active conversation.
func (d *Deps) StartChat(ctx context.Context, in StartChatInput) (domain.Conversation, error) {
	if in.TenantID == uuid.Nil || in.CustomerID == uuid.Nil {
		return domain.Conversation{}, fmt.Errorf("%w: tenant_id, customer_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	channel := in.Channel
	if channel == "" {
		channel = "web"
	}
	c := domain.Conversation{
		ID: d.newID(), TenantID: in.TenantID, CustomerID: in.CustomerID,
		AgentID: in.AgentID, TicketID: in.TicketID, Status: domain.ConversationStatusActive,
		Channel: channel, StartedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Chats.SaveConversation(ctx, c); err != nil {
		return domain.Conversation{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventChatStarted, map[string]any{
		"customerId": c.CustomerID.String(), "channel": c.Channel,
	})
	return c, nil
}

// PostMessageInput posts a chat message.
type PostMessageInput struct {
	TenantID       uuid.UUID
	ConversationID uuid.UUID
	SenderRole     string
	SenderID       *uuid.UUID
	Body           string
}

// PostMessage adds a message; creates conversation linkage is already required.
func (d *Deps) PostMessage(ctx context.Context, in PostMessageInput) (domain.Message, domain.Conversation, error) {
	if in.TenantID == uuid.Nil || in.ConversationID == uuid.Nil || strings.TrimSpace(in.Body) == "" {
		return domain.Message{}, domain.Conversation{}, fmt.Errorf("%w: tenant_id, conversation_id, body required", domain.ErrInvalidArgument)
	}
	if !domain.ValidSender(in.SenderRole) {
		return domain.Message{}, domain.Conversation{}, fmt.Errorf("%w: invalid sender_role", domain.ErrInvalidArgument)
	}
	c, err := d.Chats.GetConversation(ctx, in.TenantID, in.ConversationID)
	if err != nil {
		return domain.Message{}, domain.Conversation{}, err
	}
	if c.Status == domain.ConversationStatusEnded {
		return domain.Message{}, domain.Conversation{}, fmt.Errorf("%w: conversation ended", domain.ErrConflict)
	}
	now := d.now()
	sentiment := ""
	if d.LLM != nil {
		if s, err := d.LLM.AnalyzeSentiment(ctx, in.Body); err == nil {
			sentiment = s
		}
	}
	m := domain.Message{
		ID: d.newID(), TenantID: in.TenantID, ConversationID: c.ID,
		SenderRole: in.SenderRole, SenderID: in.SenderID,
		Body: strings.TrimSpace(in.Body), Sentiment: sentiment, CreatedAt: now,
	}
	if err := d.Chats.AddMessage(ctx, m); err != nil {
		return domain.Message{}, domain.Conversation{}, err
	}
	c.UpdatedAt = now
	_ = d.Chats.SaveConversation(ctx, c)
	return m, c, nil
}

// TransferChatInput transfers a conversation to another agent.
type TransferChatInput struct {
	TenantID       uuid.UUID
	ConversationID uuid.UUID
	ToAgentID      uuid.UUID
}

// TransferChat moves an active chat to another agent.
func (d *Deps) TransferChat(ctx context.Context, in TransferChatInput) (domain.Conversation, error) {
	c, err := d.Chats.GetConversation(ctx, in.TenantID, in.ConversationID)
	if err != nil {
		return domain.Conversation{}, err
	}
	now := d.now()
	c.TransferredFrom = c.AgentID
	c.AgentID = &in.ToAgentID
	c.Status = domain.ConversationStatusTransferred
	c.UpdatedAt = now
	if err := d.Chats.SaveConversation(ctx, c); err != nil {
		return domain.Conversation{}, err
	}
	sys := domain.Message{
		ID: d.newID(), TenantID: in.TenantID, ConversationID: c.ID,
		SenderRole: domain.SenderSystem, Body: "Conversation transferred", CreatedAt: now,
	}
	_ = d.Chats.AddMessage(ctx, sys)
	return c, nil
}

// EndChatInput ends a conversation.
type EndChatInput struct {
	TenantID       uuid.UUID
	ConversationID uuid.UUID
}

// EndChat marks a conversation ended.
func (d *Deps) EndChat(ctx context.Context, in EndChatInput) (domain.Conversation, error) {
	c, err := d.Chats.GetConversation(ctx, in.TenantID, in.ConversationID)
	if err != nil {
		return domain.Conversation{}, err
	}
	now := d.now()
	c.Status = domain.ConversationStatusEnded
	c.EndedAt = &now
	c.UpdatedAt = now
	if err := d.Chats.SaveConversation(ctx, c); err != nil {
		return domain.Conversation{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventChatEnded, nil)
	return c, nil
}
