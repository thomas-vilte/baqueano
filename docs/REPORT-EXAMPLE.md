# baqueano: an example report, built by hand

Before writing any code, I wanted to know exactly what a good report looks like. So I did the agent's job myself for one question: I searched Hacker News through the same Algolia API the agent will use, read the two threads that looked most useful, and wrote the report with the structure from [DESIGN.md](DESIGN.md#what-the-agent-delivers).

Everything here is real: every ref points to an existing item, and every problem comes from what those items say. I used only Hacker News, because Stack Exchange wasn't reachable from the environment where I built it; the report says so in "What I didn't find", as the agent would.

This document is the contract for phase 1. A report is useful if it looks like this one and passes the [checklist](#what-makes-a-report-useful) at the end.

## Contents

1. [The run](#the-run)
2. [What the model sends](#what-the-model-sends)
3. [What my code adds](#what-my-code-adds)
4. [The final report](#the-final-report)
5. [What validation rejects](#what-validation-rejects)
6. [What makes a report useful](#what-makes-a-report-useful)
7. [What building it by hand changed in the design](#what-building-it-by-hand-changed-in-the-design)

## The run

Question: *"what problems do people have with Terraform state?"* (built on 2026-10-13).

These are the tool calls the agent would make, in order:

| Turn | Call                                                              | What came back                                                                              |
|------|-------------------------------------------------------------------|---------------------------------------------------------------------------------------------|
| 1    | `search_hackernews(query: "terraform state", type: story)`        | 20 stories. Most are product launches (remote backends, scanners); two are long discussions |
| 1    | `search_hackernews(query: "terraform state lock", type: comment)` | Nothing new: the hits were the same stories                                                 |
| 2    | `read_hackernews_thread(id: 45273352)`                            | "Stategraph: Terraform state as a distributed systems problem", 61 comments                 |
| 2    | `read_hackernews_thread(id: 37809111)`                            | "A more mature take on stateless Terraform", 89 comments                                    |
| 3    | `submit_report(...)`                                              |                                                                                             |

Three turns, four tool calls, well inside the budget. The second search returning the same stories is the case the system prompt describes as "stop when results repeat".

## What the model sends

The model fills `submit_report`. It doesn't count anything and doesn't sort: it groups documents into problems and attaches refs. Refs can also appear inside the text (in workarounds, for example), and the validator checks those too.

This example uses plain refs. The current schema replaces them with `evidence`: a ref plus a short quote copied exactly from the document (see the last section). I left the example as I built it, because from where I built it I could only read paraphrases of most comments, not their exact text, and a made-up quote would defeat the point. The agent does get the exact text in every tool result, so its quotes can be checked.

```json
{
  "question": "what problems do people have with Terraform state?",
  "summary": "The most repeated complaint is that one state with one lock doesn't scale to several teams: people wait on each other, and splitting the state brings its own dependency problems. Close behind is secrets stored in plain text in the state. Refactors that destroy and recreate resources, drift that makes applies unpredictable and slow large states come up less often. Almost all the evidence comes from two Hacker News threads, so take the order as a hint, not a measurement.",
  "problems": [
    {
      "title": "One state, one lock: teams block each other",
      "description": "Teams that share a state wait on a single global lock, and Terraform doesn't run several plans or applies in parallel over the same state. Shared pieces like networking, IAM and databases make it worse as more teams touch them. An apply from an outdated checkout can also undo a teammate's change.",
      "workarounds": "Splitting the state per service or team so nobody waits on anyone else [hn:45273941] [hn:45279879]. The cost is dependencies between states and glue code to orchestrate them [hn:45273978]. One commenter says OpenTofu allows parallel plans and that Atlantis's fix didn't work [hn:45277647].",
      "refs": ["hn:45273978", "hn:45277647", "hn:45279879", "hn:45273941", "hn:37809677"]
    },
    {
      "title": "Secrets stored in plain text in the state",
      "description": "Secret values end up in the state file in plain text, including the ones the Vault provider reads. Anyone who can read the state can read them, and access can't be audited per person or the secrets rotated. Posts show exposed state buckets with live AWS keys, and state used for privilege escalation.",
      "workarounds": "Treating the whole state as sensitive, which is HashiCorp's guidance. One commenter says that isn't enough and asks for built-in encryption or hashing [hn:37810008].",
      "refs": ["hn:37809677", "hn:37810008", "hn:37831415", "hn:48263936", "hn:39543095"]
    },
    {
      "title": "Refactors destroy and recreate live resources",
      "description": "Moving code into modules or renaming resources can make Terraform destroy a resource and create a new one. One team had outages from it (a network load balancer). Merging root modules breaks on name collisions, which have to be renamed first.",
      "workarounds": "`moved` blocks, which don't accept variables. In one case, a homemade script that generates them by comparing names after a reorganization [hn:37811901].",
      "refs": ["hn:37811901", "hn:45276334"]
    },
    {
      "title": "Drift makes applies unpredictable",
      "description": "Every apply reconciles three things: what's in the cloud, what's in the code and what's recorded in the state. Provider upgrades, vendor-side changes, manual edits and drift nobody marked as ignored make applies unpredictable even when the code didn't change. Someone making a small change ends up dealing with drift in shared infrastructure they don't own.",
      "workarounds": "Splitting the state so unrelated drift stays out of your plan [hn:45279879]. One commenter prefers looking resources up by tag and creating whatever is missing [hn:45276250].",
      "refs": ["hn:45276250", "hn:45279879"]
    },
    {
      "title": "Large states slow everything down",
      "description": "As an environment grows, a single state file becomes a bottleneck: plans take longer and locks are held longer. People still want one state per environment to keep the dependencies between resources visible.",
      "workarounds": "Splitting the state, with the same trade-offs as for locking.",
      "refs": ["hn:45273809", "hn:45279879"]
    },
    {
      "title": "Losing or corrupting the state",
      "description": "The state is a single point of failure: if it gets lost or corrupted, recovering is painful.",
      "workarounds": "Not mentioned.",
      "refs": ["hn:37818343"]
    }
  ],
  "existing": [
    {
      "text": "Stategraph stores state as a graph with finer-grained locking instead of one JSON file with a global lock. Its author takes part in the thread.",
      "refs": ["hn:45273352", "hn:45273978"]
    },
    {
      "text": "Self-hosted remote backends on Cloudflare Workers, launched as Show HN projects.",
      "refs": ["hn:40964395", "hn:41458874"]
    },
    {
      "text": "Infralog tracks changes in Terraform state.",
      "refs": ["hn:43422299"]
    },
    {
      "text": "OpenTofu, which according to one commenter allows parallel plans that Terraform doesn't.",
      "refs": ["hn:45277647"]
    }
  ],
  "gaps": "I only searched Hacker News; Stack Exchange wasn't queried. All 12 comments I cite come from two threads, one about Stategraph and one about stateless Terraform, so these are mostly the problems of people who already think state is broken. I didn't find numbers on how many teams are affected or on team sizes."
}
```

## What my code adds

Between `submit_report` and the Markdown, my code does what the model is bad at or shouldn't decide:

1. **Validates refs.** Every ref, in `refs` and inside the text, has to have appeared in a tool result. Here all 18 did.
2. **Counts mentions three ways,** using the `Author` and `Thread` of each document:

| Problem                     | Documents | Authors | Threads | Note                       |
|-----------------------------|-----------|---------|---------|----------------------------|
| One state, one lock         | 5         | 5       | 2       | 1 from the thread's author |
| Secrets in plain text       | 5         | 4       | 3       | One author appears twice   |
| Refactors destroy resources | 2         | 2       | 2       |                            |
| Drift                       | 2         | 2       | 1       |                            |
| Large states                | 2         | 2       | 1       |                            |
| Losing the state            | 1         | 1       | 1       |                            |

3. **Sorts by authors,** then threads, then documents. The same person commenting twice isn't two people with the problem. Comments in the same thread aren't independent either, which is why the thread count is shown.
4. **Flags documents from the thread's author.** One of the comments in the locking problem is from the author of Stategraph, in the thread where they launch it. That's still a valid source, but the reader should know. Code detects it by comparing the comment's author with the author of the thread's root document.
5. **Renders Markdown.**

## The final report

```markdown
# Problems with Terraform state (2026-10-13)

Sources: Hacker News · 2 searches, 2 threads read · 18 documents cited

## Summary
The most repeated complaint is that one state with one lock doesn't scale to several
teams: people wait on each other, and splitting the state brings its own dependency
problems. Close behind is secrets stored in plain text in the state. Refactors that
destroy and recreate resources, drift that makes applies unpredictable and slow large
states come up less often. Almost all the evidence comes from two Hacker News threads,
so take the order as a hint, not a measurement.

## Problems

### 1. One state, one lock: teams block each other
5 documents · 5 authors · 2 threads · 1 from the thread's author

Teams that share a state wait on a single global lock, and Terraform doesn't run several
plans or applies in parallel over the same state. Shared pieces like networking, IAM and
databases make it worse as more teams touch them. An apply from an outdated checkout can
also undo a teammate's change.

**How they deal with it today:** splitting the state per service or team so nobody waits
on anyone else [hn:45273941] [hn:45279879]. The cost is dependencies between states and
glue code to orchestrate them [hn:45273978]. One commenter says OpenTofu allows parallel
plans and that Atlantis's fix didn't work [hn:45277647].

**Evidence:** [hn:45273978] [hn:45277647] [hn:45279879] [hn:45273941] [hn:37809677]

### 2. Secrets stored in plain text in the state
5 documents · 4 authors · 3 threads

Secret values end up in the state file in plain text, including the ones the Vault
provider reads. Anyone who can read the state can read them, and access can't be audited
per person or the secrets rotated. Posts show exposed state buckets with live AWS keys,
and state used for privilege escalation.

**How they deal with it today:** treating the whole state as sensitive, which is
HashiCorp's guidance. One commenter says that isn't enough and asks for built-in
encryption or hashing [hn:37810008].

**Evidence:** [hn:37809677] [hn:37810008] [hn:37831415] [hn:48263936] [hn:39543095]

### 3. Refactors destroy and recreate live resources
2 documents · 2 authors · 2 threads

Moving code into modules or renaming resources can make Terraform destroy a resource and
create a new one. One team had outages from it (a network load balancer). Merging root
modules breaks on name collisions, which have to be renamed first.

**How they deal with it today:** `moved` blocks, which don't accept variables. In one
case, a homemade script that generates them by comparing names after a reorganization
[hn:37811901].

**Evidence:** [hn:37811901] [hn:45276334]

### 4. Drift makes applies unpredictable
2 documents · 2 authors · 1 thread

Every apply reconciles three things: what's in the cloud, what's in the code and what's
recorded in the state. Provider upgrades, vendor-side changes, manual edits and drift
nobody marked as ignored make applies unpredictable even when the code didn't change.
Someone making a small change ends up dealing with drift in shared infrastructure they
don't own.

**How they deal with it today:** splitting the state so unrelated drift stays out of
your plan [hn:45279879]. One commenter prefers looking resources up by tag and creating
whatever is missing [hn:45276250].

**Evidence:** [hn:45276250] [hn:45279879]

### 5. Large states slow everything down
2 documents · 2 authors · 1 thread

As an environment grows, a single state file becomes a bottleneck: plans take longer and
locks are held longer. People still want one state per environment to keep the
dependencies between resources visible.

**How they deal with it today:** splitting the state, with the same trade-offs as for
locking.

**Evidence:** [hn:45273809] [hn:45279879]

### 6. Losing or corrupting the state
1 document · only one person mentions it

The state is a single point of failure: if it gets lost or corrupted, recovering is
painful.

**How they deal with it today:** not mentioned.

**Evidence:** [hn:37818343]

## What already exists
- Stategraph stores state as a graph with finer-grained locking instead of one JSON file
  with a global lock. Its author takes part in the thread. [hn:45273352] [hn:45273978]
- Self-hosted remote backends on Cloudflare Workers, launched as Show HN projects.
  [hn:40964395] [hn:41458874]
- Infralog tracks changes in Terraform state. [hn:43422299]
- OpenTofu, which according to one commenter allows parallel plans that Terraform
  doesn't. [hn:45277647]

## What I didn't find
I only searched Hacker News; Stack Exchange wasn't queried. All 12 comments I cite come
from two threads, one about Stategraph and one about stateless Terraform, so these are
mostly the problems of people who already think state is broken. I didn't find numbers
on how many teams are affected or on team sizes.

## Sources
Threads: A = "Stategraph: Terraform state as a distributed systems problem",
B = "A more mature take on stateless Terraform".

- [hn:45273978] lawnchair, comment in A (thread author) - https://news.ycombinator.com/item?id=45273978
- [hn:45277647] nikolay, comment in A - https://news.ycombinator.com/item?id=45277647
- [hn:45279879] Too, comment in A - https://news.ycombinator.com/item?id=45279879
- [hn:45273941] giveita, comment in A - https://news.ycombinator.com/item?id=45273941
- [hn:37809677] Hnrobert42, comment in B - https://news.ycombinator.com/item?id=37809677
- [hn:37810008] leetrout, comment in B - https://news.ycombinator.com/item?id=37810008
- [hn:37831415] Hnrobert42, comment in B - https://news.ycombinator.com/item?id=37831415
- [hn:48263936] "I built a scanner that found 41 live AWS keys in 900 Terraform state files", GeorgeWoff25 - https://news.ycombinator.com/item?id=48263936
- [hn:39543095] "Hacking Terraform State for Privilege Escalation", dagrz - https://news.ycombinator.com/item?id=39543095
- [hn:37811901] djha-skin, comment in B - https://news.ycombinator.com/item?id=37811901
- [hn:45276334] sausagefeet, comment in A - https://news.ycombinator.com/item?id=45276334
- [hn:45276250] cyberax, comment in A - https://news.ycombinator.com/item?id=45276250
- [hn:45273809] pst, comment in A - https://news.ycombinator.com/item?id=45273809
- [hn:37818343] ta1243, comment in B - https://news.ycombinator.com/item?id=37818343
- [hn:45273352] "Stategraph: Terraform state as a distributed systems problem", lawnchair - https://news.ycombinator.com/item?id=45273352
- [hn:40964395] "Show HN: I built a Terraform State remote backend on Cloudflare Workers", willswire - https://news.ycombinator.com/item?id=40964395
- [hn:41458874] "Terraform State backend using Cloudflare Workers and R2", kagitac - https://news.ycombinator.com/item?id=41458874
- [hn:43422299] "Infralog: Track Terraform state changes", hvalls - https://news.ycombinator.com/item?id=43422299
```

## What validation rejects

These variations of the same report don't make it to Markdown. Each one is a golden test in `internal/report`:

| Variation                                                  | Why it's rejected                                                           |
|------------------------------------------------------------|-----------------------------------------------------------------------------|
| A problem cites `hn:99999999`                              | That ref never appeared in a tool result                                    |
| A workaround says "see [hn:12345678]"                      | Refs inside the text are checked too, not just `refs`                       |
| A problem with an empty `refs`                             | Every problem needs evidence                                                |
| "OpenTofu has state encryption" in `existing`, with no ref | It may be true, but no document returned in this run says so                |
| An empty `gaps`                                            | "What I didn't find" is required, even if it only says the search went well |
| An evidence quote that isn't in the document's text        | The quote was paraphrased or made up; it has to be copied exactly           |

The fourth row is a real case. I know OpenTofu added state encryption, and it would answer the second problem better than anything in this report. But none of the documents I read mention it, so the report can't say it. What that tells me is that the search fell short, not that the rule is wrong: a third query, "terraform state encryption", would probably have found it. That's the kind of gap the evals have to measure.

## What makes a report useful

This is the checklist I use in phase 1 to decide whether a report counts toward the exit criterion. From phase 3 on, the judge's rubric in [EVALS.md](EVALS.md#the-judge) does the same thing at scale.

- [ ] Every problem is concrete: it says who has it and in what situation, not just "performance issues".
- [ ] Every problem is supported by at least one of its cited documents, not just linked to it, and its quotes show the problem.
- [ ] The order makes sense given the counts, and a problem with a single source says so.
- [ ] Workarounds are there when the documents mention them, and "Not mentioned" when they don't.
- [ ] "What already exists" only lists things that some document mentions.
- [ ] "What I didn't find" names the sources that weren't used or fell short, and any obvious bias in the evidence.
- [ ] After reading it, I can decide whether to keep looking into the topic or drop it.

The last item is the one that matters. The rest are there so I can tell why a report failed it.

## What building it by hand changed in the design

Doing the agent's job once showed me six things the design didn't cover. They're all applied in [DESIGN.md](DESIGN.md):

1. **Mentions are counted by documents, authors and threads, and sorted by authors.** Counting only documents would have put secrets on the same level as locking, but one person wrote two of those comments.
2. **`Document` gets a `Thread` field,** the ref of the root of its thread. Without it I can't count threads or detect the thread's author.
3. **Documents from the thread's author are flagged.** A product's author defending it in their own launch thread is a different kind of evidence than a user complaining.
4. **`Existing` items carry refs, like problems.** Before, they were plain strings, and nothing stopped the model from writing whatever it knew from training.
5. **Refs inside text fields are validated too,** not just the `refs` list.
6. **A valid ref doesn't mean a supported claim.** One comment (`hn:37809641`) looked like it was about slow plans but actually argues that state exists to avoid them. That's why the evals have a support check and a `must_not_cite` list.

One more change came out of writing the injection cases: content that contains `</document>` is escaped before it reaches the model.

And a few more came later, when I went back over what made this report good or weak. They're in [How I get good reports](DESIGN.md#how-i-get-good-reports):

- **Every piece of evidence carries an exact quote,** checked in code. It's the cheapest way to catch a ref that doesn't support its problem, which is what happened with `hn:37809641`.
- **Searches use complaint phrases and look in comments.** My first search here mostly returned launches; the complaints were in the comments.
- **The model sees how many threads it has read,** so it doesn't build a report from two conversations like this one.
- **Searches default to the last three years,** and the report shows each evidence's date.
- **The report call can use a stronger model** than the search loop, decided by evals.