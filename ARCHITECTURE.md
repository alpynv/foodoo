# Architecture

## Candidate services

- **Gateway** — authentication, routing, rate limits, request IDs.
- **Catalog** — restaurants, menus, availability; backed by Postgres and Redis.
- **Order** — creates and owns order state; transactional outbox; idempotency keys.
- **Restaurant/inventory** — accepts or rejects orders and reserves items.
- **Payment (fake)** — authorizes/captures/refunds payment; never duplicate a charge.
- **Dispatch** — courier availability, assignment, ETA, and delivery status.
- **Notification worker** — email/push simulation.
- **Search indexer** — consumes catalog events and indexes OpenSearch.

Start with fewer services: catalog, order, payment, and a notification worker. Split only when a boundary teaches something useful.

## Infrastructure

- **Postgres**: service-owned persistent data.
- **Redis**: cache, rate limiting, short-lived state.
- **Redpanda/Kafka**: durable domain-event stream.
- **RabbitMQ or NATS**: optional work queue; choose one broker initially.
- **MinIO**: restaurant images, receipts, and generated invoices.
- **OpenSearch**: restaurant/menu search.
- **Docker Compose**: initial local environment; migrate to kind later.

## Events and reliability

Example events: `OrderCreated`, `OrderAccepted`, `OrderRejected`, `PaymentAuthorized`, `PaymentFailed`, `CourierAssigned`, `OrderDelivered`.

Every event includes an event ID, schema version, order ID, correlation/request ID, timestamp, and trace context. Consumers are idempotent. The order service writes its state change and outbox record in one database transaction; an outbox worker publishes the record and retries safely.

The order workflow is a saga. For example, a payment failure after reservation emits a compensation command/event to release the reservation.
