package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type FeatureRepo interface {
	Upsert(ctx context.Context, f domain.FeatureRecord) error
	Get(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, name string, version int) (domain.FeatureRecord, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]domain.FeatureRecord, error)
}

type ModelRepo interface {
	Save(ctx context.Context, m domain.ModelCard) error
	Get(ctx context.Context, tenantID uuid.UUID, key, version string) (domain.ModelCard, error)
	GetProd(ctx context.Context, tenantID uuid.UUID, key string) (domain.ModelCard, error)
	List(ctx context.Context, tenantID uuid.UUID, key string) ([]domain.ModelCard, error)
}

type PromptRepo interface {
	Save(ctx context.Context, p domain.PromptTemplate) error
	GetActive(ctx context.Context, tenantID uuid.UUID, key, locale string) (domain.PromptTemplate, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.PromptTemplate, error)
}

type MemoryRepo interface {
	Append(ctx context.Context, m domain.ConversationMemory) error
	ListSession(ctx context.Context, tenantID, sessionID uuid.UUID, limit int) ([]domain.ConversationMemory, error)
}

type AgentRepo interface {
	SaveRun(ctx context.Context, r domain.AgentRun) error
	GetRun(ctx context.Context, tenantID, id uuid.UUID) (domain.AgentRun, error)
	ListRuns(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.AgentRun, error)
}

type AutomationRepo interface {
	SaveRule(ctx context.Context, r domain.AutomationRule) error
	ListRules(ctx context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error)
	SaveRun(ctx context.Context, r domain.AutomationRun) error
}

type DriftRepo interface {
	Save(ctx context.Context, d domain.DriftReport) error
	List(ctx context.Context, tenantID uuid.UUID, modelKey string, limit int) ([]domain.DriftReport, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// InferenceRuntime calls Python/ONNX sidecar or in-process heuristics.
type InferenceRuntime interface {
	Predict(ctx context.Context, model domain.ModelCard, features map[string]float64, inputs map[string]any) (map[string]float64, map[string]any, error)
}

// LLMProvider is a chat/completion provider.
type LLMProvider interface {
	Complete(ctx context.Context, provider, system, user string, tools []string) (content string, toolCalls []domain.ToolCall, tokensIn, tokensOut int, err error)
}

// RAGRetriever fetches grounding chunks (vector/search port).
type RAGRetriever interface {
	Retrieve(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]string, error)
}

// VectorEmbedder creates embeddings (used by search via this platform).
type VectorEmbedder interface {
	Embed(ctx context.Context, tenantID uuid.UUID, text string) ([]float64, error)
}
