# Testing and load testing

## Test layers

1. **Unit tests**: validation, prices, state transitions, idempotency, retry/backoff.
2. **Integration tests**: real Postgres, Redis, and broker via Testcontainers; test migrations, outbox publishing, and consumer behavior.
3. **End-to-end tests**: run the Compose stack and assert final state across services.
4. **Contract tests**: validate OpenAPI/gRPC contracts and versioned event schemas.

## Essential failure cases

- Duplicate `POST /orders` with the same idempotency key creates one order.
- Duplicate events do not create duplicate reservations, charges, or notifications.
- Payment rejection releases inventory.
- A worker crash after handling a message is safe when it restarts.
- Broker outage leaves outbox records to be delivered later.
- Poison messages land in a dead-letter queue with enough context to inspect/replay.
- Out-of-order events do not regress order state.

## Load testing

Use **k6** for HTTP traffic; use broker-native tools or a small producer/consumer benchmark for event throughput.

Traffic mix:

- 90% browse/search menus
- 8% create/order-status requests
- 2% checkout/payment

Scenarios: baseline, gradual ramp, short spike, multi-hour soak, and a flash sale with many simultaneous attempts for limited inventory.

Track p50/p95/p99 latency, error rate, request rate, queue depth/consumer lag, worker throughput, database pool saturation, cache hit rate, and end-to-end order completion time. Correctness matters too: no overselling, duplicate orders, or duplicate charges.

## Resilience testing

Use Docker stop/restart, resource limits, and Toxiproxy (or `tc netem`) to inject latency, disconnects, loss, slow dependencies, and consumer outages. Verify both recovery and the final business state.
