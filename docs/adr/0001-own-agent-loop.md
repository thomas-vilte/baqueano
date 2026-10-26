# ADR-0001: Write the agent loop by hand, without a framework

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

The main goal of baqueano is to learn how agents work, and to have something to show for it. The core of an agent is the loop: call the model, run the tools it asks for, send the results back, enforce budgets, decide when to stop. Frameworks hide exactly that part.

## Decision

I write the loop myself in Go, against the Messages API, through the official SDK. No agent framework.

## Options considered

### Option A: Hand-written loop

| Dimension      | Assessment                                       |
|----------------|--------------------------------------------------|
| Complexity     | Medium: a few hundred lines plus tests           |
| Cost           | No extra dependencies                            |
| Learning value | High: every step of the protocol is visible      |
| Flexibility    | Full control over budgets, memory and validation |

**Pros:** I see and test every step; budgets, memory and citation checks fit where I want them; nothing changes under me when a framework releases a new version.
**Cons:** I have to write retries, parallel tools and error handling myself; no ready-made integrations.

### Option B: An agent framework (Go or Python)

| Dimension      | Assessment                                |
|----------------|-------------------------------------------|
| Complexity     | Low to start, high when I need to bend it |
| Cost           | A large dependency with its own concepts  |
| Learning value | Low: the protocol is abstracted away      |
| Flexibility    | Limited to what the framework exposes     |

**Pros:** faster first version; integrations included.
**Cons:** hides the part I want to learn; most mature options are in Python, not Go; validation and budget rules end up as workarounds.

### Option C: Claude Agent SDK

**Pros:** a production-grade loop with tools, context and compaction already solved.
**Cons:** same problem as B, even stronger: it's someone else's loop. It's also Python and TypeScript, not Go.

## Trade-off analysis

For a product with a deadline, B or C would win. For this project the loop is the deliverable, so writing it is the point, not a cost.

## Consequences

- Easier: testing with a fake model, adding validation steps, seeing every request in traces.
- Harder: I own retries, concurrency and edge cases (cut-off answers, oversized requests).
- To revisit: if the loop grows past what I can maintain, or once the learning goal is met and I want features faster.

## Action items

1. [ ] Phase 1: loop with budgets, parallel tools with `errgroup`, `submit_report`.
2. [ ] Tests with a scripted fake model for every stop condition.
