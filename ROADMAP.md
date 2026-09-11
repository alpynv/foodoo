# Suggested roadmap

1. Create a Compose environment with one order API, Postgres, migrations, and a health endpoint.
2. Add catalog/menu browsing, then Redis caching.
3. Add an order state machine, idempotency keys, and a transactional outbox.
4. Add Redpanda/Kafka (or RabbitMQ/NATS), an event consumer, retries, and a DLQ.
5. Add fake payment and inventory/restaurant confirmation; implement compensation as a saga.
6. Add courier dispatch and WebSocket/SSE order-status updates.
7. Add MinIO for images/receipts and OpenSearch for menu search.
8. Add OpenTelemetry Collector, Prometheus, Grafana, Loki, and Tempo.
9. Add Testcontainers integration tests, Compose E2E tests, k6 load tests, and fault injection.
10. Move the stack to kind/Kubernetes only after the Compose version is understood and stable.
11. Add authentication/authorization, service credentials, rate limiting, and multi-tenant restaurant isolation.
12. Add feature flags, dynamic configuration, backward-compatible database/event migrations, and schema compatibility checks in CI.
13. Add audit trails plus reconciliation jobs that detect and repair mismatches between orders, reservations, payments, and events.
14. Practice operations: backups and restore drills, broker retention/replay, secrets management, image/dependency scanning, runbooks, and DLQ replay procedures.
15. Deliberately inject ambiguous failures (for example, payment succeeds but recording the result fails) and prove reconciliation reaches the correct final state.

Keep the first vertical slice small: create an order, persist it, publish an event, process it asynchronously, and observe the full trace.
