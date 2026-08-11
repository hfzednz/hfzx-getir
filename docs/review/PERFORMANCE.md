# Performance Report (Prompt-42)

## Applied

- Order place locking moved to Postgres advisory locks (correctness under multi-instance; latency ≈ one lock round-trip)  
- Removed dead admin mock list builders (smaller admin bundle surface)  
- Catalog/identity graceful close avoids leaked DB/Kafka writers on SIGTERM

## Observations / next

- Order saga still updates order/saga with ignored errors in some steps — retry/backoff audit pending  
- Inventory Redis holds wired earlier; inventory SQL repos still memory — hot path durability gap  
- Flutter startup/app-size profiling not run this wave  
- Kafka/OpenSearch throughput not load-tested this wave  
