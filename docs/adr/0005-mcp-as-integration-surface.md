# ADR-0005: MCP as the integration surface, in both directions

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

I want to use baqueano from Claude Code and Claude Desktop, compare Claude's research with my agent's, and let users add sources baqueano doesn't ship, without each integration becoming custom code.

## Decision

MCP is the only integration protocol:

- **As a server,** baqueano exposes its source tools, `add_document` and `run_research`. Results carry text in `content` and JSON in `structuredContent`, validated against an `outputSchema` generated from the Go types.
- **As a client** (phase 3), baqueano loads external sources as MCP servers. Their tools must return the same document schema; their refs get an `ext:` prefix and provenance `external`.
- Claude uses its own search for the open web and calls baqueano for what it can't reach. Claude doesn't push its findings into baqueano, so baqueano never cites what it didn't fetch or receive explicitly.

## Options considered

### Option A: MCP in both directions

| Dimension        | Assessment                                  |
|------------------|---------------------------------------------|
| Complexity       | Medium: server and client from the same SDK |
| Interoperability | High: any MCP client or server              |
| Isolation        | Plugins run as separate processes           |

**Pros:** one protocol; plugins in any language; a crashing plugin doesn't take the agent down.
**Cons:** baqueano can't control what an external process reaches; schema validation is the only guard.

### Option B: A REST API plus Go plugins

**Pros:** familiar.
**Cons:** custom clients for each integration; Go's `plugin` package needs the exact same toolchain and only works on some platforms.

### Option C: Merge Claude's findings into baqueano

**Pros:** one combined result.
**Cons:** baqueano would cite documents it never fetched; the validation would stop meaning anything.

## Trade-off analysis

A keeps every boundary explicit: what baqueano fetched, what a person provided and what a plugin returned are always distinguishable.

## Consequences

- Easier: using baqueano from any MCP client; adding sources without touching the core.
- Harder: versioning the MCP output schema; testing plugins.
- To revisit: OAuth for the remote server in phase 4.

## Action items

1. [ ] Phase 2: MCP server over stdio with source tools, `add_document`, `run_research`.
2. [ ] Phase 3: MCP client for external sources and a `plugin check` command that validates a plugin against the schema.
