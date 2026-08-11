package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/domain"
)

// UpsertPrompt saves a prompt template.
func (d *Deps) UpsertPrompt(ctx context.Context, p domain.PromptTemplate) (domain.PromptTemplate, error) {
	if p.TenantID == uuid.Nil || p.Key == "" || p.UserTpl == "" {
		return p, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	if p.Version <= 0 {
		p.Version = 1
	}
	p.Active = true
	p.UpdatedAt = d.now()
	return p, d.Prompts.Save(ctx, p)
}

// CompleteLLM runs prompt orchestration with RAG, tools, guardrails.
func (d *Deps) CompleteLLM(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	var out domain.LLMResponse
	if req.TenantID == uuid.Nil {
		return out, domain.ErrInvalidArgument
	}
	if blocked, reason := domain.GuardrailScan(req.UserPrompt); blocked {
		return domain.LLMResponse{ID: d.newID(), Blocked: true, BlockReason: reason}, domain.ErrGuardrailBlocked
	}

	system := "You are NEXORA AI."
	user := req.UserPrompt
	if req.PromptKey != "" {
		if tpl, err := d.Prompts.GetActive(ctx, req.TenantID, req.PromptKey, req.Locale); err == nil {
			system = tpl.System
			user = domain.RenderTemplate(tpl.UserTpl, req.Variables)
			if req.UserPrompt != "" {
				user = user + "\n" + req.UserPrompt
			}
		}
	}

	citations := []string{}
	if req.RAGQuery != "" && d.RAG != nil {
		chunks, err := d.RAG.Retrieve(ctx, req.TenantID, req.RAGQuery, 5)
		if err == nil && len(chunks) > 0 {
			system += "\nGrounding:\n" + strings.Join(chunks, "\n---\n")
			citations = chunks
		}
	}

	if req.SessionID != nil {
		_ = d.Memory.Append(ctx, domain.ConversationMemory{
			ID: d.newID(), TenantID: req.TenantID, SessionID: *req.SessionID,
			Role: "user", Content: user, CreatedAt: d.now(),
		})
		hist, _ := d.Memory.ListSession(ctx, req.TenantID, *req.SessionID, 10)
		var b strings.Builder
		for _, h := range hist {
			b.WriteString(h.Role)
			b.WriteString(": ")
			b.WriteString(h.Content)
			b.WriteString("\n")
		}
		user = b.String()
	}

	provider := req.Provider
	if provider == "" {
		provider = "mock"
	}
	content, tools, tin, tout, err := d.LLM.Complete(ctx, provider, system, user, req.Tools)
	if err != nil {
		return out, err
	}
	if blocked, reason := domain.GuardrailScan(content); blocked {
		return domain.LLMResponse{ID: d.newID(), Provider: provider, Blocked: true, BlockReason: reason}, domain.ErrGuardrailBlocked
	}

	out = domain.LLMResponse{
		ID: d.newID(), Provider: provider, Content: content, ToolCalls: tools,
		TokensIn: tin, TokensOut: tout, Citations: citations,
	}
	if req.SessionID != nil {
		_ = d.Memory.Append(ctx, domain.ConversationMemory{
			ID: d.newID(), TenantID: req.TenantID, SessionID: *req.SessionID,
			Role: "assistant", Content: content, CreatedAt: d.now(),
		})
	}
	return out, nil
}

// RunAgent executes a named enterprise agent.
func (d *Deps) RunAgent(ctx context.Context, tenantID uuid.UUID, kind, input string, sessionID *uuid.UUID) (domain.AgentRun, error) {
	var run domain.AgentRun
	if !domain.ValidAgent(kind) || tenantID == uuid.Nil {
		return run, domain.ErrInvalidArgument
	}
	steps := []domain.AgentStep{
		{Type: "thought", Content: "Selecting tools for " + kind},
	}
	// optional tool: forecast / fraud via inference
	if kind == domain.AgentForecast {
		res, err := d.Infer(ctx, domain.InferenceRequest{
			TenantID: tenantID, ModelKey: "demand_forecast",
			Features: map[string]float64{"horizon_days": 7},
			Inputs:   map[string]any{"query": input},
		})
		if err == nil {
			steps = append(steps, domain.AgentStep{Type: "tool", Name: "demand_forecast", Content: "ok"})
			steps = append(steps, domain.AgentStep{Type: "observe", Content: "predictions ready"})
			_ = res
		}
	}
	if kind == domain.AgentPricing {
		steps = append(steps, domain.AgentStep{Type: "tool", Name: "pricing_suggest", Content: "human_gated"})
	}

	llm, err := d.CompleteLLM(ctx, domain.LLMRequest{
		TenantID: tenantID, Provider: "mock", SessionID: sessionID,
		UserPrompt: input, Variables: map[string]string{"agent": kind},
		Tools: []string{"lookup_catalog", "lookup_order"},
	})
	status := "succeeded"
	output := ""
	if err != nil {
		status = "failed"
		if err == domain.ErrGuardrailBlocked {
			status = "blocked"
		}
		output = err.Error()
	} else {
		output = llm.Content
		steps = append(steps, domain.AgentStep{Type: "answer", Content: output})
		for _, tc := range llm.ToolCalls {
			steps = append(steps, domain.AgentStep{Type: "tool", Name: tc.Name, Content: "requested"})
		}
	}

	// prepend system persona via step
	steps = append([]domain.AgentStep{{Type: "thought", Name: "system", Content: domain.DefaultSystemPrompt(kind)}}, steps...)

	run = domain.AgentRun{
		ID: d.newID(), TenantID: tenantID, Kind: kind, Input: input, Output: output,
		Steps: steps, Status: status, CreatedAt: d.now(),
	}
	_ = d.Agents.SaveRun(ctx, run)
	d.emit(ctx, tenantID, run.ID, domain.EventAgentExecuted, map[string]any{
		"kind": kind, "status": status,
	})
	return run, nil
}

// UpsertAutomationRule stores a decision rule.
func (d *Deps) UpsertAutomationRule(ctx context.Context, r domain.AutomationRule) (domain.AutomationRule, error) {
	if r.TenantID == uuid.Nil || r.Name == "" || len(r.Conditions) == 0 {
		return r, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		r.ID = d.newID()
	}
	r.Enabled = true
	r.UpdatedAt = d.now()
	return r, d.Automation.SaveRule(ctx, r)
}

// EvaluateAutomation runs matching rules against a feature map.
func (d *Deps) EvaluateAutomation(ctx context.Context, tenantID uuid.UUID, features map[string]float64, approve bool) ([]domain.AutomationRun, error) {
	rules, err := d.Automation.ListRules(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	runs := make([]domain.AutomationRun, 0)
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		matched := true
		for _, c := range rule.Conditions {
			actual := features[c.Feature]
			if !domain.EvalCondition(actual, c.Op, c.Value) {
				matched = false
				break
			}
		}
		run := domain.AutomationRun{
			ID: d.newID(), TenantID: tenantID, RuleID: rule.ID, Matched: matched, CreatedAt: d.now(),
		}
		if matched {
			if rule.RequireApproval && !approve {
				run.Approved = false
				run.Result = "pending_approval"
			} else {
				run.Approved = true
				run.Result = "actions_executed"
				for _, a := range rule.Actions {
					switch a.Type {
					case "invoke_model":
						_, _ = d.Infer(ctx, domain.InferenceRequest{TenantID: tenantID, ModelKey: a.Target, Features: features})
					case "invoke_agent":
						_, _ = d.RunAgent(ctx, tenantID, a.Target, "automation:"+rule.Name, nil)
					}
				}
				d.emit(ctx, tenantID, run.ID, domain.EventAutomationTriggered, map[string]any{
					"ruleId": rule.ID.String(), "name": rule.Name,
				})
			}
		} else {
			run.Result = "no_match"
		}
		_ = d.Automation.SaveRun(ctx, run)
		runs = append(runs, run)
	}
	return runs, nil
}

// AdminStats dashboard counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	models, _ := d.Models.List(ctx, tenantID, "")
	rules, _ := d.Automation.ListRules(ctx, tenantID)
	drifts, _ := d.Drift.List(ctx, tenantID, "", 50)
	return map[string]any{
		"models": len(models), "rules": len(rules), "recentDrifts": len(drifts),
		"tenantId": tenantID.String(),
	}, nil
}
