# High error rate

## Symptoms
- Alert `HighErrorRate` firing
- Elevated 5xx in Grafana Platform Overview

## Immediate
1. Check recent Argo CD sync / canary analysis
2. If canary, halt promotion / rollback GitOps tag
3. Inspect Tempo traces for failing dependency
4. Scale healthy replicas if capacity-related

## Verify
- Error rate < 1% for 15m
- Error budget burn rate normal

## Escalate
- Page on-call SRE if revenue path (checkout/payment) affected
