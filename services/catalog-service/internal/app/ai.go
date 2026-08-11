package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
)

// DescribeProduct calls AI describe port.
func (d *Deps) DescribeProduct(ctx context.Context, tenantID, productID uuid.UUID) (ports.AIDescribeResult, error) {
	if d.AI == nil {
		return ports.AIDescribeResult{Title: "stub", Description: "AI describe stub — wire media-service + LLM adapter", Keywords: []string{"catalog", "stub"}}, nil
	}
	return d.AI.Describe(ctx, tenantID, productID)
}

// TranslateProduct calls AI translate port.
func (d *Deps) TranslateProduct(ctx context.Context, tenantID, productID uuid.UUID, lang string) (ports.AITranslateResult, error) {
	if d.AI == nil {
		return ports.AITranslateResult{Lang: lang, Title: "stub", Description: "AI translate stub"}, nil
	}
	return d.AI.Translate(ctx, tenantID, productID, lang)
}

// CategorizeProduct calls AI categorize port.
func (d *Deps) CategorizeProduct(ctx context.Context, tenantID, productID uuid.UUID) (ports.AICategorizeResult, error) {
	if d.AI == nil {
		return ports.AICategorizeResult{Confidence: 0.5}, nil
	}
	return d.AI.Categorize(ctx, tenantID, productID)
}

// QualityScoreProduct calls AI quality score port.
func (d *Deps) QualityScoreProduct(ctx context.Context, tenantID, productID uuid.UUID) (ports.AIQualityResult, error) {
	if d.AI == nil {
		return ports.AIQualityResult{Score: 0.75, Summary: "AI quality stub", Issues: []string{}}, nil
	}
	return d.AI.QualityScore(ctx, tenantID, productID)
}
