# ADR-0008: Phase 4 on AWS: Lambda, Function URL, SSM and S3, without DynamoDB

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas
**Supersedes:** [ADR-0007](0007-deploy-local-then-lambda.md)

## Context

ADR-0007 chose local Docker first and Lambda later, but left three things open or wrong: Jaeger as the local backend, session memory in DynamoDB, and "SSM or Secrets Manager". Looking at how the remote MCP server is actually used settled them. Over MCP, Claude holds the conversation, so there's no session to store. Runs take up to 2 minutes, one user, a few runs a day.

## Decision

- Local: Docker Compose with `grafana/otel-lgtm`; the CLI and stdio MCP server run on the host.
- AWS: Lambda (`provided.al2023`, arm64, zip, 512 MB, 150 s timeout) behind a Function URL, MCP over Streamable HTTP in stateless mode with JSON responses, a bearer token checked in the handler, reserved concurrency of 2.
- Secrets in SSM Parameter Store SecureString, read once at init.
- One S3 bucket for reports and eval results. No DynamoDB.

## Options considered

### Option A: Lambda with a Function URL

| Dimension   | Assessment                           |
|-------------|--------------------------------------|
| Complexity  | Low: one function, one URL           |
| Cost        | Zero at my volume (always-free tier) |
| Scalability | Scales to zero and up on its own     |
| Familiarity | High                                 |

**Pros:** no idle cost; no 29-second limit like API Gateway's HTTP APIs; one binary.
**Cons:** per-instance cache and rate limiters; cold starts; the public URL depends on my token check.

### Option B: Lambda behind API Gateway

**Pros:** throttling, WAF, more auth options.
**Cons:** HTTP APIs stop at 29 seconds; raising it only works for REST APIs and can lower account throttling.

### Option C: ECS Fargate or ECS Express Mode

**Pros:** long-lived process; no 15-minute limit; in-memory state works.
**Cons:** costs money while idle, plus a load balancer. App Runner is no longer open to new customers.

### Option D: Bedrock AgentCore Runtime

**Pros:** built for agents, pays only active CPU.
**Cons:** its own contract and ecosystem, in a project whose point is building the agent myself.

## Trade-off analysis

Idle cost and the 2-minute runs decide it: Lambda with a Function URL is the only option at zero cost without a timeout problem. DynamoDB would only buy a shared cache and rate limiter, and with reserved concurrency of 2 there are at most two warm instances. Secrets Manager's rotation doesn't apply to third-party keys I rotate by hand.

## Consequences

- Easier: zero idle cost, no session store, one binary for Compose and Lambda.
- Harder: the perimeter is my token check; streaming progress would need `RESPONSE_STREAM`.
- To revisit: DynamoDB and OAuth if more people use it; ECS Express Mode if runs outgrow Lambda.

## Action items

1. [ ] `cmd/lambda` with an in-process Function URL adapter around the MCP HTTP handler.
2. [ ] Check the account's concurrency quota before relying on reserved concurrency.
3. [ ] Check MCP client timeouts for a 2-minute tool call over HTTP.
