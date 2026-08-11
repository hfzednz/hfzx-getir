# Runbook — Redis memory high

## Symptoms

- `RedisMemoryHigh`
- Elevated latency / eviction

## Mitigate

1. Identify keyspace / big keys.
2. Flush only ephemeral caches (never session authority alone without identity fallback).
3. Scale Redis memory / cluster shard.
4. Reduce TTLs via config if safe.
5. Watch identity/session hit rate after change.
