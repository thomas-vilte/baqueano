# Architecture decision records

Each record explains one decision: the context, the options I weighed, what I chose and what it costs me. [DESIGN.md](../DESIGN.md) and [INFRA.md](../INFRA.md) describe the system as it is; these records explain why it's that way, so I can revisit a decision without redoing the whole analysis.

| ADR                                          | Decision                                                                      | Status             |
|----------------------------------------------|-------------------------------------------------------------------------------|--------------------|
| [0001](0001-own-agent-loop.md)               | Write the agent loop by hand, without a framework                             | Accepted           |
| [0002](0002-model-interface.md)              | My own `Model` interface, one provider first                                  | Accepted           |
| [0003](0003-report-through-a-tool.md)        | The report arrives through a tool and is validated in code                    | Accepted           |
| [0004](0004-sources-public-apis-only.md)     | Public APIs only; Reddit through Brave snippets and manual input              | Accepted           |
| [0005](0005-mcp-as-integration-surface.md)   | MCP as the integration surface, in both directions                            | Accepted           |
| [0006](0006-offline-evals.md)                | Offline evals with recorded responses and an LLM judge                        | Accepted           |
| [0007](0007-deploy-local-then-lambda.md)     | Local with Docker first, Lambda with a Function URL later                     | Superseded by 0008 |
| [0008](0008-phase-4-aws-setup.md)            | Phase 4 on AWS: Lambda, Function URL, SSM and S3, without DynamoDB            | Accepted           |
| [0009](0009-terraform.md)                    | Terraform for infrastructure, with S3 state and native locking                | Accepted           |
| [0010](0010-telemetry-export-and-backend.md) | OpenTelemetry straight from the SDK, Grafana Cloud in prod, otel-lgtm locally | Accepted           |
| [0011](0011-config-loading.md)               | Load configuration with the standard library and YAML, in four layers         | Accepted           |

A new record gets the next number. When a decision changes, I don't edit the old record: I write a new one that supersedes it and mark the old one as superseded.
