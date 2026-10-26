# ADR-0006: Offline evals with recorded responses and an LLM judge

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

Prompt, model and memory changes need to be measured, not judged by eye. Live sources change every day and cost quota, and the model isn't deterministic.

## Decision

A fixed set of 25 cases runs offline against recorded tool responses. Exact checks (valid refs, valid quotes, budgets, structure) run in code; coverage, support and quality are scored by Sonnet 5.5 with a written rubric, and I review a sample by hand on every run. Each configuration runs three times. Sources whose terms restrict storage (Brave) and manual input are never recorded; their cases use synthetic fixtures.

## Options considered

### Option A: Offline replay plus code checks plus LLM judge

| Dimension       | Assessment                                        |
|-----------------|---------------------------------------------------|
| Reproducibility | High                                              |
| Cost            | About USD 4 per full run with Haiku and the judge |
| Fidelity        | Medium: replays miss new queries                  |

**Pros:** comparable runs; cheap; no quota used.
**Cons:** replay misses when the model searches something new; the judge has biases.

### Option B: Live evals only

**Pros:** real data.
**Cons:** results change with the sources; costs quota; not comparable over time.

### Option C: Manual review only

**Pros:** best judgment.
**Cons:** doesn't scale past a handful of questions.

## Trade-off analysis

A gives stable numbers; the miss rate shows when fixtures are stale, and the hand-reviewed sample keeps the judge honest.

## Consequences

- Easier: comparing prompts, models and memory strategies.
- Harder: maintaining fixtures; deciding what can be committed.
- To revisit: the coverage threshold after the first full run; whether to commit recorded responses per source.

## Action items

1. [ ] Phase 1: `--record` flag.
2. [ ] Phase 3: runner, judge prompt, the 25 cases.
