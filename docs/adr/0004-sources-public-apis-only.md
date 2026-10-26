# ADR-0004: Public APIs only; Reddit through Brave snippets and manual input

**Status:** Accepted
**Date:** 2026-10-23
**Deciders:** Thomas

## Context

The best informal complaints about products are on Reddit. I asked for Reddit Data API access twice and was rejected both times under its Responsible Builder Policy, which restricts research outside its own program and sharing data with third parties such as model providers. Other sources (Hacker News, Stack Exchange, GitHub, Discourse) have public APIs. Brave Search indexes Reddit and offers an LLM Context endpoint with comments and authors.

## Decision

Built-in sources use public, documented APIs only: Hacker News, Stack Exchange, GitHub, Discourse and Brave Search. The agent never calls reddit.com and never opens a result's URL. Reddit content reaches the agent only as fragments in Brave's results, under Brave's terms, or as manual input that a person read and passed in. No cookies, no unauthenticated endpoints, no headless browsers, no third-party scraped data, no "raw content" options that fetch pages on my behalf.

## Options considered

### Option A: Public APIs, Brave for discussions, manual input for single threads

| Dimension  | Assessment                                          |
|------------|-----------------------------------------------------|
| Coverage   | Good for technical problems; partial for Reddit     |
| Risk       | Low: every source's terms are known                 |
| Durability | High: documented APIs                               |
| Cost       | Brave at USD 5 per 1,000 requests, with free credit |

**Pros:** nothing to hide in a public repo; no account at risk; stable.
**Cons:** no full Reddit threads; Brave's index lags a few days.

### Option B: Reddit through session cookies or unofficial endpoints

**Pros:** full threads.
**Cons:** against Reddit's terms after an explicit rejection; risks the account; breaks when Reddit changes anything; contradicts the rest of the design.

### Option C: Search APIs that fetch the page ("raw content")

**Pros:** more text per result.
**Cons:** it's scraping with an intermediary.

## Trade-off analysis

The loss is real but smaller than it looks: for developer tools, GitHub issues and Hacker News are often more precise than Reddit, and Brave's comments with authors cover much of the informal side.

## Consequences

- Easier: a repo I can show; sources I can keep running.
- Harder: a specific new thread needs manual input.
- To revisit: if Reddit offers an access path that fits this use case.

## Action items

1. [ ] Phase 2: Brave client with `search_web` and `search_discussions`, `record: false`.
2. [ ] Phase 2: manual input with provenance `provided`.
3. [ ] Confirm which Brave plans include storage rights.
