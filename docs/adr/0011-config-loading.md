# ADR-0011: Load configuration with the standard library and YAML, in four layers

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

The config covers models per stage, budgets, sources and cache. It has four fixed layers (defaults, YAML file, environment, flags), has to fail at startup on any mistake, and its effective value has to be hashed into `config_hash` so a result says what produced it.

## Decision

A `Config` struct, `go.yaml.in/yaml/v3` with `KnownFields(true)`, and the layers merged by hand in `config.Load`. `Validate()` returns every problem at once with `errors.Join`. Environment variables are `BAQUEANO_` plus the YAML path. Secrets live in a separate struct and never enter the hash. On Lambda the prod config and the prompts are embedded with `//go:embed`.

## Options considered

### Option A: Standard library and YAML

| Dimension   | Assessment             |
|-------------|------------------------|
| Complexity  | Low: about 150 lines   |
| Cost        | One dependency         |
| Scalability | Enough for four layers |
| Familiarity | Total: it's my code    |

**Pros:** the precedence is readable in one function; easy to test with an injected `getenv`.
**Cons:** I write the env and flag mapping myself.

### Option B: koanf

**Pros:** modular, few dependencies, the best option if layers multiply.
**Cons:** more than I need today.

### Option C: viper

**Pros:** popular, many sources.
**Cons:** lowercases keys, global state by default, a large dependency tree, hard-to-predict merging.

## Trade-off analysis

With four fixed layers, a library hides the one thing I want visible: which value wins.

## Consequences

- Easier: a typo in the YAML fails at startup; `config_hash` is deterministic.
- Harder: every new field needs its env and flag mapping (a table test catches missing ones).
- To revisit: koanf if a remote config source ever appears.

## Action items

1. [ ] `internal/config` with `Defaults`, `Load`, `Validate` and `Hash`.
2. [ ] Table tests for precedence and validation.
