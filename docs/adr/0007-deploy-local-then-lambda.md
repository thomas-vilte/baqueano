# ADR-0007: Local with Docker first, Lambda with a Function URL later

**Status:** Superseded by [ADR-0008](0008-phase-4-aws-setup.md)
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

One user, a few questions a day, runs of up to two minutes. It has to be cheap when idle and reachable from Claude when I'm away from my machine.

## Decision

Phases 1 to 3 run locally: a static binary in a distroless image, plus a Compose file with an OpenTelemetry Collector and Jaeger, and MCP over stdio. Phase 4 deploys to AWS Lambda (`provided.al2023`, arm64) behind a Function URL, with MCP over Streamable HTTP in stateless mode, session memory in DynamoDB with expiry, secrets in SSM or Secrets Manager, and OAuth in front.

## Options considered

### Option A: Lambda with a Function URL

| Dimension         | Assessment                                       |
|-------------------|--------------------------------------------------|
| Cost when idle    | Zero                                             |
| Cost at my volume | Inside the free tier                             |
| Complexity        | Medium: stateless server, external session store |

**Pros:** pay per use; no servers to manage; Go starts fast.
**Cons:** stateless by design; a 15-minute ceiling; in-memory cache only lives per instance.

### Option B: Fargate or App Runner

**Pros:** long-lived process; in-memory state works.
**Cons:** costs money while idle.

### Option C: Stay local

**Pros:** simplest.
**Cons:** not reachable from claude.ai on the web.

## Trade-off analysis

At this volume, idle cost decides. Lambda's stateless model forces session memory out of the process, which I need to design anyway.

## Consequences

- Easier: zero idle cost; secrets managed by AWS.
- Harder: the cache and the rate limiters are per instance, not shared.
- To revisit: a shared cache or rate limiter if usage grows.

## Action items

1. [ ] Phase 4: Lambda handler, DynamoDB session store, OAuth, reserved concurrency, AWS Budgets alarm.
