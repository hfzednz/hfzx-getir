package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/domain"
)

func (d *Deps) UpsertPolicy(ctx context.Context, p domain.SecurityPolicy) (domain.SecurityPolicy, error) {
	if p.TenantID == uuid.Nil || p.Key == "" || !domain.ValidPolicyKind(p.Kind) {
		return p, domain.ErrInvalidArgument
	}
	p.Key = domain.NormalizeKey(p.Key)
	now := d.now()
	existing, err := d.Policies.GetByKey(ctx, p.TenantID, p.Key)
	if err == nil {
		p.ID = existing.ID
		p.Version = existing.Version + 1
		p.CreatedAt = existing.CreatedAt
	} else {
		if p.ID == uuid.Nil {
			p.ID = d.newID()
		}
		p.Version = 1
		p.CreatedAt = now
	}
	p.Enabled = true
	p.UpdatedAt = now
	if p.Rego == "" {
		p.Rego = "package nexora\ndefault allow = false\n"
	}
	return p, d.Policies.Save(ctx, p)
}

func (d *Deps) EvaluatePolicy(ctx context.Context, tenantID uuid.UUID, policyKey, subject, action, resource string, input map[string]any) (domain.PolicyDecision, error) {
	if tenantID == uuid.Nil || policyKey == "" || subject == "" || action == "" {
		return domain.PolicyDecision{}, domain.ErrInvalidArgument
	}
	policyKey = domain.NormalizeKey(policyKey)
	pol, err := d.Policies.GetByKey(ctx, tenantID, policyKey)
	if err != nil {
		return domain.PolicyDecision{}, err
	}
	if !pol.Enabled {
		return domain.PolicyDecision{}, domain.ErrPolicyDenied
	}
	if input == nil {
		input = map[string]any{}
	}
	input["subject"] = subject
	input["action"] = action
	input["resource"] = resource

	allow := false
	reason := "default_deny"
	risk := 0.0
	if d.OPA != nil {
		a, r, err := d.OPA.Evaluate(ctx, pol.Rego, input)
		if err != nil {
			return domain.PolicyDecision{}, err
		}
		allow, reason = a, r
	} else {
		// Fail-safe allow for read when no OPA and policy kind access with action read
		if pol.Kind == domain.PolicyAccess && action == "read" {
			allow, reason = true, "fallback_read_allow"
		}
	}
	if idTrust, ok := input["identityTrust"].(float64); ok {
		devTrust, _ := input["deviceTrust"].(float64)
		ctxRisk, _ := input["contextRisk"].(float64)
		trust := domain.AdaptiveTrustScore(idTrust, devTrust, ctxRisk)
		risk = 1 - trust
		if trust < 0.4 {
			allow = false
			reason = "adaptive_trust_too_low"
		}
	}
	dec := domain.PolicyDecision{
		ID: d.newID(), TenantID: tenantID, PolicyKey: policyKey,
		Subject: subject, Action: action, Resource: resource,
		Allow: allow, Reason: reason, RiskScore: risk, Context: input, EvaluatedAt: d.now(),
	}
	_ = d.Policies.SaveDecision(ctx, dec)
	if !allow {
		d.emit(ctx, tenantID, dec.ID, domain.EventPolicyViolated, map[string]any{
			"policyKey": policyKey, "subject": subject, "action": action, "reason": reason,
		})
		_, _ = d.RecordAudit(ctx, domain.AuditEvent{
			TenantID: tenantID, ActorID: subject, ActorType: "user", Action: "policy.evaluate",
			ResourceType: "policy", ResourceID: policyKey, Outcome: "denied",
		})
	}
	return dec, nil
}

func (d *Deps) RecordAudit(ctx context.Context, e domain.AuditEvent) (domain.AuditEvent, error) {
	if e.TenantID == uuid.Nil || e.Action == "" || e.ActorID == "" {
		return e, domain.ErrInvalidArgument
	}
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = d.now()
	}
	if e.Outcome == "" {
		e.Outcome = "success"
	}
	prev, _ := d.Audits.LastHash(ctx, e.TenantID)
	e.PrevHash = prev
	payload := e.ActorID + "|" + e.Action + "|" + e.ResourceType + "|" + e.ResourceID + "|" + e.Outcome
	e.Hash = domain.ChainHash(prev, payload)
	return e, d.Audits.Append(ctx, e)
}

func (d *Deps) RegisterSecret(ctx context.Context, s domain.SecretMeta) (domain.SecretMeta, error) {
	if s.TenantID == uuid.Nil || s.Name == "" || s.VaultPath == "" {
		return s, domain.ErrInvalidArgument
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	s.Name = domain.NormalizeKey(s.Name)
	if s.Kind == "" {
		s.Kind = "api_key"
	}
	s.Status = "active"
	s.Rotatable = true
	s.Version = 1
	s.CreatedAt = d.now()
	return s, d.Secrets.Save(ctx, s)
}

func (d *Deps) RotateSecret(ctx context.Context, tenantID, secretID uuid.UUID) (domain.SecretMeta, error) {
	s, err := d.Secrets.Get(ctx, tenantID, secretID)
	if err != nil {
		return s, err
	}
	if !s.Rotatable || s.Status == "revoked" {
		return s, domain.ErrSecretNotRotatable
	}
	s.Status = "rotating"
	_ = d.Secrets.Save(ctx, s)
	ver := s.Version + 1
	if d.Vault != nil {
		v, err := d.Vault.Rotate(ctx, s.VaultPath)
		if err != nil {
			return s, err
		}
		ver = v
	}
	now := d.now()
	s.Version = ver
	s.LastRotated = &now
	s.Status = "active"
	if err := d.Secrets.Save(ctx, s); err != nil {
		return s, err
	}
	d.emit(ctx, tenantID, s.ID, domain.EventSecretRotated, map[string]any{
		"name": s.Name, "version": s.Version, "vaultPath": s.VaultPath,
	})
	_, _ = d.RecordAudit(ctx, domain.AuditEvent{
		TenantID: tenantID, ActorID: "system", ActorType: "system", Action: "secret.rotate",
		ResourceType: "secret", ResourceID: s.ID.String(), Outcome: "success",
	})
	return s, nil
}

func (d *Deps) RenewCertificate(ctx context.Context, tenantID, secretID uuid.UUID) (domain.SecretMeta, error) {
	s, err := d.Secrets.Get(ctx, tenantID, secretID)
	if err != nil {
		return s, err
	}
	if s.Kind != "cert" {
		return s, domain.ErrInvalidArgument
	}
	var exp time.Time
	if d.Vault != nil {
		exp, err = d.Vault.RenewCertificate(ctx, s.VaultPath)
		if err != nil {
			return s, err
		}
	} else {
		exp = d.now().Add(90 * 24 * time.Hour)
	}
	s.ExpiresAt = &exp
	s.Version++
	now := d.now()
	s.LastRotated = &now
	_ = d.Secrets.Save(ctx, s)
	d.emit(ctx, tenantID, s.ID, domain.EventCertificateRenewed, map[string]any{
		"name": s.Name, "expiresAt": exp,
	})
	return s, nil
}

func (d *Deps) IngestThreat(ctx context.Context, tenantID uuid.UUID, subject string, features map[string]float64) (domain.ThreatAlert, error) {
	if tenantID == uuid.Nil || subject == "" {
		return domain.ThreatAlert{}, domain.ErrInvalidArgument
	}
	kind, score := domain.DetectThreatKind(features)
	if d.Fraud != nil {
		if fs, err := d.Fraud.Score(ctx, tenantID, features); err == nil && fs > score {
			score = fs
		}
	}
	t := domain.ThreatAlert{
		ID: d.newID(), TenantID: tenantID, Kind: kind, Severity: domain.SeverityFromScore(score),
		Subject: subject, Score: score, Indicators: map[string]any{"features": features},
		Status: "open", CreatedAt: d.now(),
	}
	if err := d.Threats.Save(ctx, t); err != nil {
		return t, err
	}
	d.emit(ctx, tenantID, t.ID, domain.EventThreatDetected, map[string]any{
		"kind": kind, "severity": t.Severity, "score": score, "subject": subject,
	})
	d.emit(ctx, tenantID, t.ID, domain.EventSecurityAlertCreated, map[string]any{
		"threatId": t.ID.String(), "severity": t.Severity,
	})
	return t, nil
}

func (d *Deps) IngestFinding(ctx context.Context, f domain.ScanFinding) (domain.ScanFinding, error) {
	if f.TenantID == uuid.Nil || f.Source == "" || f.Title == "" {
		return f, domain.ErrInvalidArgument
	}
	if f.ID == uuid.Nil {
		f.ID = d.newID()
	}
	if !domain.ValidSeverity(f.Severity) {
		f.Severity = "medium"
	}
	f.Status = "open"
	f.DetectedAt = d.now()
	return f, d.Vulns.Save(ctx, f)
}

func (d *Deps) OpenIncident(ctx context.Context, i domain.Incident) (domain.Incident, error) {
	if i.TenantID == uuid.Nil || i.Title == "" {
		return i, domain.ErrInvalidArgument
	}
	if i.ID == uuid.Nil {
		i.ID = d.newID()
	}
	if !domain.ValidSeverity(i.Severity) {
		i.Severity = "medium"
	}
	i.Status = "open"
	i.OpenedAt = d.now()
	i.Timeline = []domain.IncidentEvent{{At: i.OpenedAt, Actor: "system", Message: "incident opened"}}
	if i.PlaybookKey == "" {
		i.PlaybookKey = "default_ir"
	}
	if d.SOAR != nil {
		runID, _ := d.SOAR.RunPlaybook(ctx, i.TenantID, i.PlaybookKey, map[string]any{"title": i.Title})
		i.Timeline = append(i.Timeline, domain.IncidentEvent{
			At: d.now(), Actor: "soar", Message: "playbook started " + runID,
		})
	}
	if err := d.Incidents.Save(ctx, i); err != nil {
		return i, err
	}
	d.emit(ctx, i.TenantID, i.ID, domain.EventIncidentOpened, map[string]any{
		"severity": i.Severity, "title": i.Title,
	})
	return i, nil
}

func (d *Deps) CloseIncident(ctx context.Context, tenantID, incidentID uuid.UUID, postmortem string) (domain.Incident, error) {
	i, err := d.Incidents.Get(ctx, tenantID, incidentID)
	if err != nil {
		return i, err
	}
	if i.Status == "closed" {
		return i, domain.ErrIllegalTransition
	}
	now := d.now()
	i.Status = "closed"
	i.ClosedAt = &now
	i.Postmortem = postmortem
	i.Timeline = append(i.Timeline, domain.IncidentEvent{At: now, Actor: "system", Message: "incident closed"})
	if err := d.Incidents.Save(ctx, i); err != nil {
		return i, err
	}
	d.emit(ctx, tenantID, i.ID, domain.EventIncidentClosed, map[string]any{"severity": i.Severity})
	return i, nil
}

func (d *Deps) UpsertControl(ctx context.Context, c domain.ComplianceControl) (domain.ComplianceControl, error) {
	if c.TenantID == uuid.Nil || c.Framework == "" || c.ControlID == "" {
		return c, domain.ErrInvalidArgument
	}
	if c.ID == uuid.Nil {
		c.ID = d.newID()
	}
	if c.Status == "" {
		c.Status = "not_started"
	}
	c.Framework = domain.NormalizeKey(c.Framework)
	c.UpdatedAt = d.now()
	return c, d.Compliance.SaveControl(ctx, c)
}

func (d *Deps) AddEvidence(ctx context.Context, e domain.ComplianceEvidence) (domain.ComplianceEvidence, error) {
	if e.TenantID == uuid.Nil || e.ControlID == uuid.Nil || e.Title == "" {
		return e, domain.ErrInvalidArgument
	}
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	e.CollectedAt = d.now()
	if err := d.Compliance.SaveEvidence(ctx, e); err != nil {
		return e, err
	}
	return e, nil
}

func (d *Deps) RunComplianceAudit(ctx context.Context, tenantID uuid.UUID, framework string) (domain.ComplianceAuditRun, error) {
	if tenantID == uuid.Nil || framework == "" {
		return domain.ComplianceAuditRun{}, domain.ErrInvalidArgument
	}
	framework = domain.NormalizeKey(framework)
	controls, err := d.Compliance.ListControls(ctx, tenantID, framework)
	if err != nil {
		return domain.ComplianceAuditRun{}, err
	}
	score, gaps := domain.ComplianceScore(controls)
	now := d.now()
	run := domain.ComplianceAuditRun{
		ID: d.newID(), TenantID: tenantID, Framework: framework,
		Score: score, Gaps: gaps, Status: "completed", StartedAt: now, CompletedAt: &now,
	}
	if err := d.Compliance.SaveAuditRun(ctx, run); err != nil {
		return run, err
	}
	d.emit(ctx, tenantID, run.ID, domain.EventComplianceAuditCompleted, map[string]any{
		"framework": framework, "score": score, "gaps": gaps,
	})
	return run, nil
}

func (d *Deps) UpsertDataAsset(ctx context.Context, a domain.DataAsset) (domain.DataAsset, error) {
	if a.TenantID == uuid.Nil || a.Name == "" || !domain.ValidClassification(a.Classification) {
		return a, domain.ErrInvalidArgument
	}
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	a.CreatedAt = d.now()
	return a, d.DataGov.SaveAsset(ctx, a)
}

func (d *Deps) CreatePrivacyRequest(ctx context.Context, p domain.PrivacyRequest) (domain.PrivacyRequest, error) {
	if p.TenantID == uuid.Nil || p.SubjectRef == "" {
		return p, domain.ErrInvalidArgument
	}
	switch p.Kind {
	case "export", "erase", "consent_withdraw":
	default:
		return p, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	p.Status = "open"
	p.CreatedAt = d.now()
	return p, d.DataGov.SavePrivacy(ctx, p)
}

func (d *Deps) CompletePrivacyRequest(ctx context.Context, tenantID, id uuid.UUID) (domain.PrivacyRequest, error) {
	p, err := d.DataGov.GetPrivacy(ctx, tenantID, id)
	if err != nil {
		return p, err
	}
	if p.Status == "completed" {
		return p, domain.ErrIllegalTransition
	}
	now := d.now()
	p.Status = "completed"
	p.CompletedAt = &now
	_ = d.DataGov.SavePrivacy(ctx, p)
	_, _ = d.RecordAudit(ctx, domain.AuditEvent{
		TenantID: tenantID, ActorID: "privacy", ActorType: "system",
		Action: "privacy." + p.Kind, ResourceType: "subject", ResourceID: p.SubjectRef, Outcome: "success",
	})
	return p, nil
}

func (d *Deps) UpsertRisk(ctx context.Context, r domain.RiskItem) (domain.RiskItem, error) {
	if r.TenantID == uuid.Nil || r.Title == "" || r.Category == "" {
		return r, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		r.ID = d.newID()
	}
	r.Score = domain.ComputeRiskScore(r.Likelihood, r.Impact)
	if r.Status == "" {
		r.Status = "open"
	}
	now := d.now()
	r.UpdatedAt = now
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	return r, d.Risks.Save(ctx, r)
}

func (d *Deps) RequestAccess(ctx context.Context, a domain.AccessRequest) (domain.AccessRequest, error) {
	if a.TenantID == uuid.Nil || a.RequesterID == "" || a.Resource == "" {
		return a, domain.ErrInvalidArgument
	}
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	if a.TTLMinutes <= 0 {
		a.TTLMinutes = 60
	}
	a.Status = "pending"
	a.CreatedAt = d.now()
	return a, d.Access.Save(ctx, a)
}

func (d *Deps) DecideAccess(ctx context.Context, tenantID, id uuid.UUID, approve bool) (domain.AccessRequest, error) {
	a, err := d.Access.Get(ctx, tenantID, id)
	if err != nil {
		return a, err
	}
	if a.Status != "pending" {
		return a, domain.ErrIllegalTransition
	}
	now := d.now()
	a.DecidedAt = &now
	if approve {
		a.Status = "approved"
		exp := now.Add(time.Duration(a.TTLMinutes) * time.Minute)
		a.ExpiresAt = &exp
	} else {
		a.Status = "denied"
	}
	return a, d.Access.Save(ctx, a)
}

func (d *Deps) UpsertDeviceTrust(ctx context.Context, dev domain.DeviceTrust) (domain.DeviceTrust, error) {
	if dev.TenantID == uuid.Nil || dev.DeviceID == "" {
		return dev, domain.ErrInvalidArgument
	}
	if existing, err := d.Devices.GetByDevice(ctx, dev.TenantID, dev.DeviceID); err == nil {
		dev.ID = existing.ID
	} else if dev.ID == uuid.Nil {
		dev.ID = d.newID()
	}
	score := 1.0
	if !dev.Attested {
		score -= 0.3
	}
	if dev.Rooted || dev.Jailbroken {
		score -= 0.4
	}
	if dev.Tampered {
		score -= 0.5
	}
	if score < 0 {
		score = 0
	}
	dev.TrustScore = score
	dev.LastSeenAt = d.now()
	return dev, d.Devices.Save(ctx, dev)
}

func (d *Deps) CheckAIPrompt(ctx context.Context, tenantID uuid.UUID, modelKey, prompt string) (domain.AISecurityEvent, error) {
	if tenantID == uuid.Nil || prompt == "" {
		return domain.AISecurityEvent{}, domain.ErrInvalidArgument
	}
	score := domain.PromptInjectionScore(prompt)
	blocked := score >= 0.35
	if d.AIGuard != nil {
		if b, s, err := d.AIGuard.ValidatePrompt(ctx, tenantID, modelKey, prompt); err == nil {
			if b {
				blocked = true
			}
			if s > score {
				score = s
			}
		}
	}
	e := domain.AISecurityEvent{
		ID: d.newID(), TenantID: tenantID, ModelKey: modelKey,
		PromptHash: domain.ChainHash("", prompt), Kind: "prompt_injection",
		Blocked: blocked, Score: score, CreatedAt: d.now(),
	}
	_ = d.AISec.Save(ctx, e)
	if blocked {
		d.emit(ctx, tenantID, e.ID, domain.EventSecurityAlertCreated, map[string]any{
			"kind": "ai_prompt_injection", "score": score,
		})
	}
	return e, nil
}

func (d *Deps) RecordFraudSignal(ctx context.Context, s domain.FraudSignal) (domain.FraudSignal, error) {
	if s.TenantID == uuid.Nil || s.Subject == "" {
		return s, domain.ErrInvalidArgument
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	if s.Features == nil {
		s.Features = map[string]float64{}
	}
	if d.Fraud != nil {
		if sc, err := d.Fraud.Score(ctx, s.TenantID, s.Features); err == nil {
			s.Score = sc
		}
	}
	s.CreatedAt = d.now()
	return s, d.FraudSigs.Save(ctx, s)
}

func (d *Deps) ZeroTrustEvaluate(ctx context.Context, tenantID uuid.UUID, subject, deviceID string, contextRisk float64) (map[string]any, error) {
	idTrust := 0.5
	if d.Identity != nil {
		if t, err := d.Identity.IdentityTrust(ctx, tenantID, subject); err == nil {
			idTrust = t
		}
	}
	devTrust := 0.5
	if deviceID != "" {
		if dev, err := d.Devices.GetByDevice(ctx, tenantID, deviceID); err == nil {
			devTrust = dev.TrustScore
		}
	}
	score := domain.AdaptiveTrustScore(idTrust, devTrust, contextRisk)
	return map[string]any{
		"trustScore": score, "identityTrust": idTrust, "deviceTrust": devTrust,
		"contextRisk": contextRisk, "allowHint": score >= 0.5,
	}, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	threats, _ := d.Threats.List(ctx, tenantID, "open")
	incidents, _ := d.Incidents.List(ctx, tenantID, "open")
	vulns, _ := d.Vulns.List(ctx, tenantID, "open")
	risks, _ := d.Risks.List(ctx, tenantID)
	pending, _ := d.Access.ListPending(ctx, tenantID)
	var riskSum int
	for _, r := range risks {
		riskSum += r.Score
	}
	avgRisk := 0.0
	if len(risks) > 0 {
		avgRisk = float64(riskSum) / float64(len(risks))
	}
	return map[string]any{
		"openThreats": len(threats), "openIncidents": len(incidents),
		"openVulns": len(vulns), "pendingAccess": len(pending),
		"avgRiskScore": avgRisk, "tenantId": tenantID.String(),
		"postureHint": fmt.Sprintf("threats=%d incidents=%d", len(threats), len(incidents)),
	}, nil
}
