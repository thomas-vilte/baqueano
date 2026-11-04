# Reviewing a pull request

How I review a task's pull request, and what I ask the reviewing agent to do. The agent that wrote the code never reviews it: the review starts from the spec, not from the author's explanation.

## Inputs

- The task card in [PLAN.md](PLAN.md) and the doc sections it links to.
- [AGENTS.md](../AGENTS.md).
- The diff against `main`, read in full, tests included.

The PR description is a claim to verify, not evidence. Every "done" gets checked against the code or a command.

## Steps

1. **Run it.** On the PR branch: `make fmt lint test vuln`, plus `make replay` from T17 on. `make fmt` must leave no diff. Paste the output in the review.
2. **"Done when", one by one.** For each line, name the test or the code that proves it. If no test proves it, it isn't done.
3. **Scope.** Every changed file belongs to the task's scope. Changes outside it are a finding, even if they're improvements.
4. **The checklist below.**
5. **Verdict.** One of: `approve`, `changes requested` (with the list), or `task is wrong` (the spec or the plan doesn't hold up, and I need to fix it before any more code).

## Checklist

**Spec**
- [ ] The behavior matches the linked sections of DESIGN.md, INFRA.md or EVALS.md, and the "What this plan settles" table in PLAN.md.
- [ ] Every decision the spec didn't cover is listed in the PR description.
- [ ] The public types from T04 didn't change, or the task says they can.

**Tests**
- [ ] The tests fail without the change. Spot-check at least one by breaking the code it covers.
- [ ] No test was weakened, skipped or deleted to make CI pass. Check every modified `_test.go` and every golden file.
- [ ] Failure paths are tested, not only the happy one: timeouts, 429, 5xx, broken input, cancellation.
- [ ] No real network, no `time.Sleep`; time-dependent code uses `synctest`.
- [ ] Golden files changed only where the task changes the output, and the diff makes sense line by line.

**Go**
- [ ] `context.Context` is passed and respected; nothing blocks after cancellation.
- [ ] Errors are wrapped with context and classified with `apierr`; no string comparisons, no `panic`, no ignored errors without a comment.
- [ ] No global state, no `init()` with side effects, no global OTel providers.
- [ ] Goroutines have an owner and an end; no leaks on error paths.
- [ ] The import rules in AGENTS.md hold (`go list -deps` if in doubt).
- [ ] Nothing in `internal/` writes to stdout.

**Safety**
- [ ] No secrets, keys or `.env` in code, fixtures or logs; URLs with `key=` are redacted.
- [ ] Third-party text (comments, posts) never reaches logs or metric attributes.
- [ ] New dependencies are on the allowed list in AGENTS.md, or justified in the PR.

**Telemetry** (from T15 on)
- [ ] New spans and attributes match INFRA.md; no ids, URLs or questions as metric attributes.

## Findings

Each finding has a file and line, what's wrong, why it matters, and what would fix it. Severity:

- **blocker**: wrong behavior, a missing "Done when", a weakened test, a secret, a broken import rule.
- **should fix**: works, but it will cost later (unclear errors, a missing failure test, scope creep).
- **nit**: style or naming. Optional, and never the reason to block.

## Limits

- Two rounds of `changes requested` on the same task means the task is too big or the spec is unclear. I split it or fix the plan instead of a third round.
- The reviewer approves; I merge. T04 and anything that changes a public type also gets my own read.
