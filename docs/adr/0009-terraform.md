# ADR-0009: Terraform for infrastructure, with S3 state and native locking

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

Phase 4 needs a Lambda, a Function URL, IAM roles, a log group, a bucket, a budget and an OIDC trust for GitHub Actions. I want it reproducible, reviewed in pull requests, and in a tool I'll meet again at work.

## Decision

Terraform 1.11 or later, written so OpenTofu runs it too. State in S3 with `use_lockfile = true`, no DynamoDB lock table. One `prod` environment plus local. CI runs `plan` on pull requests and `apply` on `main`, with OIDC roles and no stored AWS keys. The Lambda zip is built in the same workflow and deployed by Terraform through `source_code_hash`.

## Options considered

### Option A: Terraform

| Dimension   | Assessment                            |
|-------------|---------------------------------------|
| Complexity  | Medium                                |
| Cost        | Free CLI                              |
| Scalability | High                                  |
| Familiarity | The most common IaC in job posts; HCL |

**Pros:** industry standard; native S3 locking since 1.11; OpenTofu as an exit.
**Cons:** HCL, not Go; BSL license (doesn't affect me as a user).

### Option B: AWS CDK in Go

**Pros:** my language.
**Cons:** CloudFormation underneath; most examples in TypeScript.

### Option C: AWS SAM

**Pros:** the fastest path for a single Lambda.
**Cons:** falls short for OIDC, budgets and anything beyond Lambda; less transferable.

### Option D: Pulumi in Go

**Pros:** real Go.
**Cons:** less standard; another state service to learn.

## Trade-off analysis

The point is learning the tool I'll meet most, and Terraform with S3 locking now needs only one bucket for state. CDK would let me write Go but teaches CloudFormation; SAM would be quicker today and limiting tomorrow.

## Consequences

- Easier: infra changes reviewed as plans in pull requests; no long-lived keys.
- Harder: secrets values stay outside Terraform (created with `aws ssm put-parameter`) so they never land in state.
- To revisit: a separate code deploy if I ever ship code many times a day.

## Action items

1. [ ] `infra/bootstrap`: state bucket, OIDC provider, plan and deploy roles.
2. [ ] `infra/modules`: `lambda_mcp`, `artifacts_bucket`, `observability`.
3. [ ] `infra.yml` workflow with a required reviewer on the `prod` environment.
