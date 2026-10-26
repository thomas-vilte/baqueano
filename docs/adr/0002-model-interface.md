# ADR-0002: My own Model interface, one provider first

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

The loop needs a model. SDK types change between versions, providers differ in how they represent tool calls and errors, and I want to compare models in evals. At the same time, supporting several providers from day one doubles the work before anything runs.

## Decision

The loop depends on a small `Model` interface I own (`Complete(ctx, Request) (Response, error)`). Phase 1 has a single adapter, for Anthropic, using `anthropic-sdk-go`. A second provider comes in phase 3, once the evals can measure the difference. Model choice is per stage: one model for research and one for the report, both Haiku 5.5 in phase 1.

## Options considered

### Option A: Own interface, one adapter now

| Dimension   | Assessment                            |
|-------------|---------------------------------------|
| Complexity  | Low: one interface, one adapter       |
| Cost        | Small mapping code                    |
| Flexibility | High: new providers and fakes plug in |

**Pros:** testable with a fake model; SDK changes stay inside the adapter; per-stage models are just two values of the same interface.
**Cons:** a bit of mapping code; the interface may need changes when the second provider arrives.

### Option B: Use the SDK types directly in the loop

**Pros:** less code at first.
**Cons:** the loop depends on one provider's types; tests need the real SDK or heavy mocks; switching provider means rewriting the loop.

### Option C: A multi-provider library

**Pros:** many providers at once.
**Cons:** hides the differences I want to understand (errors, parallel calls, schema support); another large dependency.

## Trade-off analysis

A is a small cost now that pays off in tests from the first day and in evals later. C would save time on providers I don't need yet.

## Consequences

- Easier: fake models in tests, per-stage models, provider comparisons in evals.
- Harder: each adapter maps errors and usage itself.
- To revisit: the interface shape after the second adapter exists.

## Action items

1. [ ] Phase 1: `internal/model/anthropic` with error mapping to `ErrAuth`, `ErrRateLimited`, `ErrUnavailable`, `ErrBadRequest`.
2. [ ] Decide who owns retries: the SDK retries some errors by default, so either turn that off and retry in the adapter, or keep the SDK's and don't add mine. Never both.
3. [ ] Phase 3: a second adapter and a comparison in evals.
