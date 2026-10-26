# AGENTS.md

Instructions for coding agents (Codex, Claude Code) working on baqueano: a research agent in Go that finds problems people report online and returns a report with verifiable citations.

## Where the spec is

Read in this order, and only what your task needs:

1. **Your task card in [docs/PLAN.md](docs/PLAN.md).** Scope, dependencies and "Done when". The table "What this plan settles" overrides any older snippet in the other docs.
2. **The sections the card links to** in [docs/DESIGN.md](docs/DESIGN.md), [docs/INFRA.md](docs/INFRA.md), [docs/EVALS.md](docs/EVALS.md) and [docs/REPORT-EXAMPLE.md](docs/REPORT-EXAMPLE.md).
3. **[docs/adr/](docs/adr/README.md)** when you need to know why something was decided.

When the prose and a code snippet in the docs disagree, the prose wins. The snippets show the design, not finished code. When the docs and your task disagree, or the task can't be done as written, stop and explain it in the pull request. Don't invent a third version.

## Commands

```bash
make fmt      # gofmt + goimports
make lint     # golangci-lint run
make test     # go test -race ./...
make vuln     # govulncheck ./...
make replay   # offline replay of recorded runs (from T17 on)
make build    # binaries into bin/
```

A pull request is ready when `make fmt lint test vuln` passes with no changes left behind.

## How the code is written

- **`context.Context` first** in every function that does I/O or can block. Respect cancellation.
- **Small interfaces, defined where they're used.** Accept interfaces, return structs.
- **Nothing global.** No package-level state, no `init()` with side effects, no global OTel providers. Config, logger, tracer and clients come through constructors.
- **Errors:** wrap with `%w` and context (`fmt.Errorf("hackernews: read %s: %w", ref, err)`). Classify with the sentinels in `internal/apierr` and check with `errors.Is`. Never compare error strings. No `panic` in library code.
- **A tool failure is not an agent failure.** It becomes a tool result with `is_error: true`.
- **Retries live only in the source HTTP client and the model adapter.** Never in the loop. The Anthropic SDK's own retries stay at 0.
- **Concurrency:** `errgroup` with `SetLimit`; each goroutine writes to its own slice index instead of sharing a mutex when it can.
- **Logs:** `log/slog`, always with context (`InfoContext`), never with third-party text (comments, posts) or secrets.
- **stdout belongs to the output.** The CLI prints the report there, and the MCP server speaks JSON-RPC there. Everything else goes to stderr or OTLP. No `fmt.Print` in `internal/`.
- **Dependencies point inward, to the interfaces.** `internal/agent` never imports a concrete source or `internal/model/anthropic`; it only knows `Model`, `Tool` and `Memory`. Sources import `internal/tools` and never `internal/agent`. The model adapter imports `internal/agent` to implement `Model`. Only `cmd/` wires concrete pieces together (see "Architecture" in DESIGN.md).
- Comments explain why, not what. Exported identifiers have doc comments.

## Tests

- **No real network in `go test`.** Use `httptest.Server`, an in-memory `RoundTripper`, or a fake `Model`. Tests that hit real APIs go behind `//go:build live` and don't run in normal CI.
- **Everything passes with `-race`.**
- **Table-driven tests** for parsers, validation and mappings.
- **Time:** `testing/synctest` for timeouts, retries, backoff and the circuit breaker. No `time.Sleep` in tests.
- **Golden files** in `testdata/*.golden`, refreshed with `go test ./... -update`.
- **Fixtures** are synthetic or recorded with keys redacted. Never a real key, token or `.env` in `testdata/`.

## Dependencies

Allowed without asking:

- `github.com/anthropics/anthropic-sdk-go`
- `github.com/modelcontextprotocol/go-sdk` (phase 2)
- `go.opentelemetry.io/otel/...` and `go.opentelemetry.io/contrib/...`
- `go.yaml.in/yaml/v3`
- `golang.org/x/sync`, `golang.org/x/time`, `golang.org/x/text`

Anything else: propose it in the pull request with the reason and what it would replace. No web frameworks, DI containers, config libraries, mocking frameworks or assertion libraries. The standard library and hand-written fakes are enough.

## Boundaries

- Never commit secrets, `.env`, recorded responses with keys, or `bin/`.
- Don't edit files under `docs/` except to tick your task's box in `docs/PLAN.md`. If a doc is wrong, say so in the pull request.
- Don't change the public types from T04 (`internal/tools`, `internal/agent`, `internal/report`, `internal/apierr`) outside a task that says so. If you need a change, explain why in the pull request.
- Stay inside your task's scope. No drive-by refactors in other packages.

## Pull requests

- Branch `tNN-short-name`, title `TNN: what it does`.
- The description has the task's "Done when" list, each line checked or explained, plus anything you decided that the spec didn't cover.
- Small commits with messages in the imperative ("add hackernews search tool").
