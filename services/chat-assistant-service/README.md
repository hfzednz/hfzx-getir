# chat-assistant-service

Placeholder for a future dedicated AI contact-center / bot orchestration service.

**Current ownership (PROMPT-22):** AI intent detection, KB-grounded reply drafting, sentiment, summarization, and low-confidence / negative-sentiment escalation live in **`crm-service`** behind the `LLMClient` port (`github.com/nexora/crm-service`).

Per the NEXORA constitution / Master Blueprint §7, bot orchestration **may** later split here from `crm-service`. Until then:

- Prefer extending `crm-service` `internal/app` AIAssist + mock/real `LLMClient` adapters
- Do not duplicate ticket/SLA/Customer360 ownership in this folder
- This README exists so PROMPT-22 and service maps can point at the intended future split

See `../crm-service/ARCHITECTURE.md` and `../crm-service/README.md`.
