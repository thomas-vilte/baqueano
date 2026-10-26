# ADR-0003: The report arrives through a tool and is validated in code

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

The output only matters if I can trust it. Models invent citations, stretch what a source says and count badly. Free text can't be checked reliably.

## Decision

The model delivers the report by calling a `submit_report` tool with a fixed schema. My code then validates it before anything is shown: every ref must have appeared in a tool result, every quote must appear word for word in that document's text, every problem and every "already exists" item needs evidence, and "what I didn't find" can't be empty. Mentions (documents, authors, threads) are counted in code, never by the model.

## Options considered

### Option A: Structured tool output plus validation in code

| Dimension   | Assessment                                  |
|-------------|---------------------------------------------|
| Complexity  | Medium: schema, validator, golden tests     |
| Reliability | High: invented refs and quotes are rejected |
| Cost        | No extra model calls                        |

**Pros:** deterministic checks; numbers I can trust; the same structure feeds CLI, MCP and evals.
**Cons:** a rigid schema; the model sometimes needs a second try.

### Option B: Free-text report, parsed afterwards

**Pros:** more natural writing.
**Cons:** parsing is fragile; checks are best effort; counts come from the model.

### Option C: Free text plus an LLM reviewer

**Pros:** catches some unsupported claims.
**Cons:** another model call per question; the reviewer can be wrong too; nothing is guaranteed.

## Trade-off analysis

A moves the checks that can be exact (refs, quotes, counts) into code, and leaves only judgment calls (does this quote support this problem?) for the evals.

## Consequences

- Easier: trusting citations; comparing runs; rendering the same report anywhere.
- Harder: changing the schema means updating validator, renderer, evals and MCP schema together.
- To revisit: whether quotes need fuzzy matching for sources that alter whitespace or quotes.

## Action items

1. [ ] Phase 1: schema, validator and golden tests with the example report.
2. [ ] Add `schema_version` to the report so MCP consumers can tell versions apart.
