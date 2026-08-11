# FEATURES — crm-service

| Feature | Status | Notes |
|---------|--------|-------|
| CreateTicket + idempotency key | done | Same key returns existing ticket |
| AssignTicket | done | open/pending/reopened → in_progress |
| AddNote | done | Internal notes + events |
| Escalate | done | Raises priority + escalation row |
| Resolve / Close / Reopen | done | Close only from resolved |
| MergeTickets | done | Source closed into target |
| StartChat / PostMessage / Transfer / End | done | Sender roles customer\|agent\|ai\|system |
| AIAssist | done | KB + MockLLM; escalate on low conf / negative |
| UpsertArticle / PublishArticle / SearchKB | done | draft → published |
| CreateCase / UpdateCase | done | Refund type → RefundRequest port only |
| SubmitCSAT / SubmitNPS | done | CSAT 1–5, NPS 0–10 |
| GetCustomer360 | done | Profile + orders read stubs |
| EvaluateSLA / BreachEscalation | done | Marks breach; optional escalate |
| AdminStats | done | Tenant counters |
| Outbox + EventPublisher | done | Memory + stub Kafka |
| HTTP `/v1/crm/...` NEXORA errors | done | X-Tenant-Id, X-Nexora-User |
| OpenAPI + proto | done | Codegen not wired |
| Postgres migrations | done | Memory default in DevMode |
| Notification ownership | out of scope | Port only |
| Payment/refund execution | out of scope | RefundRequest only |
| Profile SoT | out of scope | Read aggregation |
