# Optimization Report (Prompt-42)

## Applied

1. Consolidated order outbound HTTP clients (one code path)  
2. Correct optimistic concurrency (avoids silent lost updates — correctness over micro-optimization)  
3. Removed dead admin mock code  
4. Resource cleanup on shutdown (catalog, identity, payment publisher)

## Not changed (by design)

- Architecture, ports, Kafka topics, service boundaries  
- Domain money model and error envelope  

## Candidates for later waves

- Shared connection pools / HTTP transport across BFF clients  
- Outbox publisher batching  
- Flutter rebuild audits (`const`, selectors)  
