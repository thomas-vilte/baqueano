# baqueano

A research agent in Go that finds real problems people report online and returns a cited report.

You ask it something like *"what problems do people have with Terraform state?"*. It searches communities where people describe what goes wrong for them (Hacker News, Stack Exchange, and later GitHub and Discourse forums), reads the threads that look relevant, and gives back a report: the problems it found, how many documents mention each one, how people work around them today, and a link to every source.

I'm building it for two reasons: to check product and project ideas before I spend time on them, and to learn how to build agents from scratch instead of only using them.

> **Status:** design phase. The full design is in [docs/DESIGN.md](docs/DESIGN.md). Code for phase 1 is next.

## How it works

![The agent loop](docs/diagrams/01-agent-loop.png)

The model doesn't run anything and doesn't remember anything between calls. It only asks for tools; my code runs them and sends the results back. The agent is that back-and-forth in a loop:

1. Send the question, the system prompt and the tool definitions to the Claude API.
2. If the model asks for tools (`stop_reason: tool_use`), run them in parallel and send back the results.
3. Repeat until the model hands in the report through a `submit_report` tool, or a budget runs out (8 turns, 2 minutes, 90k input tokens).
4. Check that every citation in the report points to a document the tools actually returned, count mentions per problem (documents, authors and threads), and render Markdown. There's [a full example](docs/REPORT-EXAMPLE.md) built by hand with real data.

![Components](docs/diagrams/02-architecture.png)

## What makes it different

- **My own loop, no agent framework.** Written against a small `Model` interface, so it's testable with a fake model and the model can change without touching the loop.
- **Verifiable citations.** Every document has a stable ref like `hn:41234567`. A ref that never appeared in a tool result is an error, not a footnote, and every piece of evidence carries a short quote that code checks against the document's actual text.
- **Mentions are counted in code, not by the model.** That number decides what's at the top of the report.
- **Hard budgets** on turns, tokens, time and cost.
- **Context management:** tools return excerpts, old results get shrunk, and compaction is the last resort.
- **OpenTelemetry traces, metrics and logs** with the GenAI semantic conventions: one span per model call and per tool call, cost per run, errors per source. Grafana locally (`otel-lgtm`) and in prod (Grafana Cloud).
- **Evals:** a fixed question set, run offline against recorded responses, to compare prompts, models and memory strategies.
- **Read-only and polite:** public APIs only, no scraping, a 1-hour in-memory cache, per-source rate limits, nothing used for training.

## Sources

| Source | Access | Phase |
| --- | --- | --- |
| Hacker News (Algolia API) | No key | 1 |
| Stack Exchange API v2.3 | Optional free key | 1 |
| GitHub Issues and Discussions | Personal token | 2 |
| Brave Search (web and LLM Context) | API key | 2 |
| Discourse forums | Each forum's public JSON | 2 |
| Manual input (a page I read and pass in) | None | 2 |
| Reddit | Access rejected; discussions partly covered through Brave | — |

Endpoints, filters and limits for each one are in [docs/DESIGN.md](docs/DESIGN.md#appendix-sources-in-detail).

## Planned usage

Not implemented yet; this is the target interface.

```bash
export ANTHROPIC_API_KEY=...

# First run: write a config with only Hacker News on, then check everything works
go run ./cmd/agent init
go run ./cmd/agent doctor

# Ask a question, get a Markdown report
go run ./cmd/agent "what problems do people have with Terraform state?"

# Add a page you read yourself as evidence (phase 2)
go run ./cmd/agent --input notes/thread.txt "what problems do people describe here?"

# Or expose the tools, and the agent itself as run_research,
# to Claude Desktop / Claude Code over MCP (stdio)
go run ./cmd/mcp
```

From Claude, the split is simple: Claude uses its own web search for the open web and calls baqueano's tools for the sources it can't reach well on its own. Every document says whether a tool fetched it or a person provided it. More in [Using it with Claude](docs/DESIGN.md#using-it-with-claude).

Configuration lives in `config.yaml` (model, budgets, enabled sources). Secrets only come from environment variables.

## Stack

Go · Claude API (`anthropic-sdk-go`, Haiku 5.5 by default; other providers plug in through the `Model` interface) · MCP Go SDK · OpenTelemetry · Grafana · Docker · AWS Lambda, SSM and S3 with Terraform (phase 4) · GitHub Actions

## Roadmap

![Phases](docs/diagrams/05-phases.png)

| Phase | Scope | Done when |
| --- | --- | --- |
| 1. MVP | Loop, budgets, Hacker News, Stack Exchange, `submit_report`, session memory, traces | The 5 phase 1 eval cases give useful reports with 100% valid citations |
| 2. Sources | MCP server, GitHub, Brave, manual input, Discourse | Each source's terms read and respected |
| 3. Evals and memory | 25-question eval set, LLM judge, injection cases, working notes, a second model provider | Stable evals across runs |
| 4. Deploy | Lambda with a Function URL, MCP over HTTP, Terraform, Grafana Cloud | Public eval results |

## Docs

- [docs/DESIGN.md](docs/DESIGN.md): the full design, with every decision explained. It starts from what an agent is and covers the loop, system prompt, tools, Go code, memory, security, observability, tests, evals and deploy.
- [docs/INFRA.md](docs/INFRA.md): where it runs and how I operate it: Docker locally, the AWS setup and its cost, Terraform, how configuration loads, every span, metric and log, the test strategy and CI.
- [docs/REPORT-EXAMPLE.md](docs/REPORT-EXAMPLE.md): a report I built by hand with real Hacker News data, from the `submit_report` input to the final Markdown. It's the target for phase 1.
- [docs/EVALS.md](docs/EVALS.md): the 25 eval cases, how they're scored, the judge's rubric and how recorded responses are replayed.
- [docs/PLAN.md](docs/PLAN.md): phase 1 split into tasks, one pull request each, with dependencies and done criteria. [AGENTS.md](AGENTS.md) has the rules for the coding agents that work on it.
- [docs/adr/](docs/adr/README.md): one record per key decision, with the options I weighed and what each choice costs.

## License

[Apache License 2.0](LICENSE).
