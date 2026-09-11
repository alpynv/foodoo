# Observability

Use OpenTelemetry everywhere and correlate metrics, logs, and traces with `request_id`, `order_id`, `event_id`, and W3C `traceparent`.

## Local stack

```text
services -> OTLP -> OpenTelemetry Collector
                    |- Prometheus -> Grafana (metrics)
                    |- Tempo or Jaeger -> Grafana (traces)
                    `- Loki -> Grafana (logs)
```

## Metrics

Expose HTTP request/error/latency metrics, queue depth or consumer lag, worker retries and throughput, database query/pool metrics, Redis hit rate, and business metrics such as orders/minute, payment failures, rejected orders, and delivery time.

## Logs

Emit structured JSON logs. Include service name, level, message, timestamp, order/request/trace IDs, and safe operational fields. Never log credentials, payment data, tokens, or other secrets.

## Traces

Create spans for inbound requests, outbound calls, database operations, message publication, and message consumption. Propagate trace context through broker headers so one trace follows an order across async workers.

## Useful alerts

- elevated error rate or p95 checkout latency
- growing consumer lag or queue depth
- non-empty dead-letter queue
- no event processing while orders arrive
- database pool near exhaustion
- failed compensation/release after a payment failure

A debugging workflow: find an `order_id` in Grafana logs, open its trace through the `trace_id`, identify the failed or slow span, then compare queue and dependency metrics.
