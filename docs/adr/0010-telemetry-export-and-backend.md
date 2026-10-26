# ADR-0010: OpenTelemetry straight from the SDK, Grafana Cloud in prod, otel-lgtm locally

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

Every run produces traces, metrics and logs. Locally I want to see them seconds after a run. On Lambda an instance can freeze right after returning, so whatever hasn't been exported is lost. I want the same code everywhere, with only environment variables changing, and a backend that costs nothing at my volume (about 15 MB of traces a month).

## Decision

- The Go SDK exports over OTLP/HTTP with protobuf, configured by the standard `OTEL_*` variables.
- On Lambda there's no collector: the handler calls `ForceFlush` on the tracer, meter and logger providers, with a 2 to 3 second timeout, before returning.
- Prod backend: Grafana Cloud Free. Local backend: `grafana/otel-lgtm` in Docker Compose.
- Spans are written by hand with the GenAI semantic conventions; all names live in `internal/telemetry/genai`.

## Options considered

### Option A: SDK export and ForceFlush

| Dimension   | Assessment                            |
|-------------|---------------------------------------|
| Complexity  | Low                                   |
| Cost        | The flush adds to the billed duration |
| Cold start  | Nothing extra                         |
| Familiarity | High                                  |

**Pros:** one process; same code locally and on Lambda.
**Cons:** a slow backend lengthens the invocation; redaction has to be in code.

### Option B: Collector as a Lambda extension (upstream layer or ADOT)

**Pros:** central processing; the `decouple` processor lets the function return before the export.
**Cons:** another process to start; the billed duration still includes the export; the upstream layer is dropping processors from its default build; no Go auto-instrumentation to gain.

### Backends

| Backend                | Why not, or why                                                              |
|------------------------|------------------------------------------------------------------------------|
| **Grafana Cloud Free** | Picked: 10k series, 50 GB of logs and traces, 14 days; same stack as local   |
| Honeycomb Free         | Great traces, but no classic logs or PromQL                                  |
| X-Ray and CloudWatch   | The OTLP traces endpoint requires SigV4, which breaks "only env vars change" |
| Self-hosted            | A server to run                                                              |

## Trade-off analysis

A collector earns its place with many services or central processing. With one service and one backend, the only gain would be not waiting for a flush of a few hundred milliseconds in a 2-minute run.

## Consequences

- Easier: one telemetry setup; dashboards built locally load into Grafana Cloud.
- Harder: the GenAI conventions are still in Development, so names can change; I pin a version and test the names.
- To revisit: a collector if a second service appears; Phoenix or Langfuse as a second destination for eval traces in phase 3.

## Action items

1. [ ] `internal/telemetry` with providers, `ForceFlush` and `Shutdown` helpers, and the redacting `RoundTripper`.
2. [ ] A test that checks metric attributes against an allowlist.
3. [ ] Dashboards in `deploy/grafana/dashboards/` and the `ErrAuth` and daily cost alerts.
