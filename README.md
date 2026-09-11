# Foodoo

A local-first food-delivery system built to learn modern distributed-systems patterns.

## Product flow

1. A customer browses restaurants and menus.
2. They submit an order with an idempotency key.
3. The order service persists the order and an outbox event atomically.
4. Restaurant/inventory confirms the order.
5. A fake payment service authorizes payment.
6. Dispatch assigns a courier; status updates are delivered in real time.
7. Notification workers send order updates.

Failures must be safe: retries do not duplicate orders or charges; a rejected payment releases the reservation; permanently failed work reaches a dead-letter queue.

## Learning goals

- independently deployable services and service-owned data
- synchronous HTTP/gRPC plus asynchronous events and work queues
- transactional outbox, idempotency, retries, DLQs, and saga compensation
- caching, object storage, search, realtime updates, and observability
- local Compose first; local Kubernetes (kind) later

See [ARCHITECTURE.md](ARCHITECTURE.md), [TESTING.md](TESTING.md), [OBSERVABILITY.md](OBSERVABILITY.md), and [ROADMAP.md](ROADMAP.md).
