# The Senior Engineer's Template

> **What this is:** a reusable, fill-in-the-blanks version of [`SENIOR-PLAYBOOK.md`](SENIOR-PLAYBOOK.md). It doesn't depend on any language, stack or domain. Copy it into any repo and use it to start a new project or to add a feature to an existing one.
>
> **How it differs from the playbook:** the playbook explains the reasoning using this password manager as the worked example. This template gives you the same thinking as an **inner monologue** (what a senior is saying to themselves at each step), followed by blanks you fill in. When a step's *why* isn't obvious, follow the "Deeper" link back to the playbook.

---

## Table of contents

- [0. How to use this](#0-how-to-use-this)
- [1. The always-on mental loop](#1-the-always-on-mental-loop)
- [Track A: Starting a new project](#track-a-starting-a-new-project)
  - [A1. Understand the problem](#a1-understand-the-problem)
  - [A2. What must not go wrong](#a2-what-must-not-go-wrong)
  - [A3. Constraints and technology](#a3-constraints-and-technology)
  - [A4. Architecture shape](#a4-architecture-shape)
  - [A5. Contracts first](#a5-contracts-first)
  - [A6. Ready to code?](#a6-ready-to-code)
- [Track B: Adding a feature](#track-b-adding-a-feature)
  - [B1. Clarify the ask](#b1-clarify-the-ask)
  - [B2. Read before you write](#b2-read-before-you-write)
  - [B3. Blast radius and compatibility](#b3-blast-radius-and-compatibility)
  - [B4. Risk and one-way doors](#b4-risk-and-one-way-doors)
  - [B5. The four questions and the test list](#b5-the-four-questions-and-the-test-list)
  - [B6. Slice into PRs](#b6-slice-into-prs)
  - [B7. The build loop](#b7-the-build-loop)
  - [B8. Self-review](#b8-self-review)
  - [B9. Ship and verify](#b9-ship-and-verify)
- [Track C: Done and operating](#track-c-done-and-operating)
- [Appendix 1: Mini templates](#appendix-1-mini-templates)
- [Appendix 2: When I'm stuck](#appendix-2-when-im-stuck)
- [Appendix 3: Anti-patterns to catch in yourself](#appendix-3-anti-patterns-to-catch-in-yourself)

---

## 0. How to use this

1. Copy this file into the target repo, e.g. as `docs/design/<project-or-feature>.md`.
2. Work through the steps that apply (see the sizing table below). Fill in every `✍️` blank. Short answers are fine. Writing the answer down is what forces the thinking.
3. Delete the steps you skipped, and keep the filled-in file next to the code. It becomes your design doc.

### Sizing: how much of this to do

Thinking effort should match how big and how irreversible the work is. Don't write a design doc for a typo fix.

| Kind of work | Do these steps | Typical time on paper |
|---|---|---|
| Bug fix or small tweak | B1, B2, B5, B7, B8, B9 | 5-15 minutes |
| Feature in an existing project | All of Track B | 30 minutes to 2 hours |
| Feature that changes a contract (API, schema, file format) | Track B, plus A5 for the changed contract | 1-4 hours |
| New project | Track A, then Track B for every feature, then Track C before calling it done | A few days at most for a solo project |

### Block format

Every step uses the same layout:

- **Goal:** what the step achieves, in one line.
- **🧠 Inner monologue:** what a senior is thinking at this point, in first person. Read it slowly. The questions in it are the actual skill.
- **✍️ Fill in:** blanks to answer. The *italic text* is an example of a good answer. Replace it with yours.
- **🚩 Junior trap:** the most common mistake at this step.
- **✅ Exit check:** how you know the step is done.

### The one rule behind all of it

> **Spend thinking in proportion to how hard the decision is to undo.**
>
> A **one-way door** (data format, public API, security model, ID type, anything users or other systems will depend on) deserves hours and a written record. A **two-way door** (function names, folder layout, which helper library) deserves seconds. Juniors tend to get this backwards. [Deeper: playbook 1.2](SENIOR-PLAYBOOK.md#12-one-way-doors-and-two-way-doors)

---

## 1. The always-on mental loop

Seniors don't run this as a checklist. They run it constantly, for a whole project, a feature, or a single function:

```
   ┌─► 1. What is this FOR?           (the real goal, not the requested solution)
   │   2. What's the RISKIEST part?    (do that first; nervousness is a good detector)
   │   3. Is this decision REVERSIBLE? (one-way door → slow down and write it down)
   │   4. What must NEVER happen?      (invariants → these become tests)
   │   5. What's the SMALLEST correct step?
   │   6. How will I KNOW it worked?   (a test, a green CI tick, a log line; not a feeling)
   └── 7. What did I learn that changes an earlier answer? → go back and update it on paper
```

### The mindset principles, one line each

| # | Principle | In one line |
|---|---|---|
| 1 | Order by risk, not ease | Do the thing most likely to sink the project first, not the thing you already know how to do. |
| 2 | One-way vs two-way doors | Think hard about irreversible decisions, and move fast on reversible ones. |
| 3 | Walking skeleton, then vertical slices | Get all the plumbing running end-to-end first, then add one thin, complete behavior at a time. |
| 4 | Work → right → fast | In risky domains (security, money, data), "works" already includes "right". |
| 5 | Write things down | Writing is thinking. A vague idea can't survive being written down. |
| 6 | Boring tech, few dependencies | Every dependency and every new technology is a cost you pay every day. |
| 7 | Optimize for the reader and for change | Ask "Will I understand this in six months?" and "How many files change if X changes?" |
| 8 | Make the wrong thing hard | If you catch yourself writing "remember to…", ask how the design could enforce it instead. |

[Deeper: playbook Part 1](SENIOR-PLAYBOOK.md#part-1-the-senior-mindset)

---

## Track A: Starting a new project

Do these in order, but expect to loop back. If A5 reveals a gap in A2, update A2.

### A1. Understand the problem

**Goal:** explain on one page what you're building, for whom, what success looks like, and what you are *not* building.

**🧠 Inner monologue**
> "Before I touch the editor, what does this thing promise? If I can't say it in one sentence, I'm not ready to design anything."
>
> "Who is actually going to use this, and what do they do today instead? If the answer is 'nothing, it's fine', why build it?"
>
> "What would make a user drop this immediately? That's what I must protect above everything else."
>
> "Every feature I say no to now is weeks I don't spend later, and complexity I never have to debug or secure. What can I cut?"
>
> "The core promise has consequences. What does it rule *out*? I want to find those in the first hour, not after someone builds a feature that can't work."
>
> "What's the real purpose of this project? Learning, portfolio, a business, a favor for a friend? That changes what 'done' means."

**✍️ Fill in: problem statement**

```markdown
## What
✍️ *One paragraph. A CLI tool that syncs X between Y and Z.*

## For whom
✍️ *Individuals who self-host; recruiters reading the code.*

## Core promise (one sentence)
✍️ *"If the server is stolen, the attacker learns nothing useful."*

## Consequences of that promise
✍️ *No password recovery is possible. The server can't search user data.*

## Must have (v1)
✍️ *3-6 bullets. The minimum that's actually useful.*

## Not in v1 (maybe later)
✍️ *Sharing, web UI, mobile app.*

## Never
✍️ *Things the core promise makes impossible, or that you'll never want.*

## Constraints
✍️ *One developer, ~1-2 h/day, runs locally, must be free to host.*

## Success looks like
✍️ *A stranger can run it with one command and understand why it's safe from the README.*
```

**🚩 Junior trap:** starting from the tutorial shape ("every app has a `User` with CRUD") before knowing what this product promises.

**✅ Exit check**
- [ ] The problem statement fits on one page.
- [ ] I can list what's out of scope without hesitating.
- [ ] I've written down at least one consequence of the core promise.

[Deeper: playbook Phase A](SENIOR-PLAYBOOK.md#phase-a-understand-the-problem-no-code)

---

### A2. What must not go wrong

**Goal:** know what you're protecting, from whom or from what, and let that shape the architecture.

**🧠 Inner monologue**
> "What are the assets here? Data, money, uptime, someone's privacy, someone's reputation?"
>
> "Who or what could hurt them? An attacker, yes. But also a buggy client, a dead disk, two requests arriving at the same moment, or me running the wrong migration at midnight."
>
> "What's the single question that decides the architecture? For a password manager it's 'if the DB is stolen, what can be read?'. What's the equivalent here?"
>
> "Where are the trust boundaries? Where does data cross from something I control to something I don't? Everything arriving across a boundary is untrusted."
>
> "I won't solve every risk. Which ones am I deliberately accepting, and why? If I write it down, it's a decision. If I don't, it's an oversight."
>
> "Requirements should come from these risks, not from a list of features most apps have."

**✍️ Fill in**

The deciding question for this project: ✍️ *"If the database is stolen, what can be read?" / "If a payment request is sent twice, what happens?"*

Assets: ✍️ *user data, account access, the audit log, availability during business hours*

Risk table (attackers *and* accidents):

| # | Threat (who or what) | What it can do | Worst outcome | Mitigation | Residual risk (accepted, and why) |
|---|---|---|---|---|---|
| R1 | ✍️ *DB thief* | *Copies the DB file* | *Reads all user data* | *Encrypt client-side* | *Weak passwords can still be brute-forced offline* |
| R2 | ✍️ *Another user* | *Has a valid account* | *Reads someone else's data* | *Every query scoped by the caller's ID* | *Bugs in scoping; tests per endpoint* |
| R3 | ✍️ *Disk failure* | — | *All data lost* | *Backups, plus a tested restore* | *Up to 24 h of data loss* |
| R4 | ✍️ *Concurrent edits* | *Two devices save at once* | *Silent overwrite* | *Revision check, 409 on mismatch* | — |

Trust boundary sketch:

```
✍️  TRUSTED: ________________     │ boundary │    UNTRUSTED: ________________
    (e.g. user's device)          │          │    (e.g. network, server, DB)
```

**🚩 Junior trap:** treating security or reliability as a feature for later ("we'll add encryption in sprint 4"). Many of these are one-way doors that reshape the schema and the API.

**✅ Exit check**
- [ ] The deciding question is answered, and the answer is written down.
- [ ] The risk table exists, with accepted residual risks.
- [ ] I can draw the trust boundary from memory.

[Deeper: playbook Phase B](SENIOR-PLAYBOOK.md#phase-b-find-the-one-thing-that-must-not-go-wrong)

---

### A3. Constraints and technology

**Goal:** pick the tools from the constraints, not from habit or hype, and write down why.

**🧠 Inner monologue**
> "How big does this actually need to be? I want numbers, not 'it should scale'."
>
> "What happens if it's down for an hour? What if data is lost? Those two answers are usually very different, and they tell me where to spend effort."
>
> "What do I already know well? A familiar boring tool beats an exciting one I'll be debugging at night."
>
> "For each dependency: what does it save me, and what does it cost me? A library that saves me from writing a security-sensitive parser is worth it. A library that wraps one stdlib call isn't."
>
> "What's the simplest architecture that meets the requirements? Every extra service is something to run, secure, monitor and upgrade."
>
> "Under what condition would I change my mind? If I write that down, the decision becomes conditional instead of permanent."

**✍️ Fill in: non-functional requirements (use numbers)**

| NFR | Target | Notes |
|---|---|---|
| Scale (users, data size, req/s) | ✍️ *< 10 users, < 100k rows, < 10 req/s* | |
| Availability | ✍️ *An hour down is annoying, not critical* | |
| Durability | ✍️ *Data loss is catastrophic → backups required* | |
| Latency | ✍️ *Nothing time-critical; must not feel broken* | |
| Security / privacy | ✍️ *See A2* | |
| Operability (who runs it, how) | ✍️ *One person, Docker Compose* | |
| Team and time | ✍️ *Solo, 1-2 h/day* | |

**✍️ Fill in: technology choices**

| Choice | Constraint that justifies it | Trade-off I accept | Door type |
|---|---|---|---|
| ✍️ *Language* | | | |
| ✍️ *Database* | | | |
| ✍️ *Transport (REST / gRPC / CLI)* | | | |
| ✍️ *Framework / router* | | | *Two-way: pick fast* |
| ✍️ *Deployment* | | | |

Things I deliberately did **not** choose, and why: ✍️ *No message queue: nothing is asynchronous yet.*

Choices that need an ADR (one-way-ish): ✍️ *database, auth scheme, ID type* (template: [Appendix 1](#appendix-1-mini-templates))

**🚩 Junior trap:** choosing the stack "real companies use" (microservices, Kubernetes, Kafka) for a project with no such requirement.

**✅ Exit check**
- [ ] NFRs are written down with numbers.
- [ ] Every major choice has a one-line justification tied to a constraint.
- [ ] One-way-ish choices have an ADR with a "Revisit when" line.

[Deeper: playbook Phase C](SENIOR-PLAYBOOK.md#phase-c-constraints-non-functional-requirements-and-choosing-technology)

---

### A4. Architecture shape

**Goal:** organize the code so each part has one job, changes stay local, and the important rules are hard to break.

**🧠 Inner monologue**
> "I'm not going to start with a pattern name. What are the different *reasons* this code will change? Each independent reason gets its own place."
>
> "For each part: what must it know, and, more importantly, what must it NOT know? A layer that doesn't know about HTTP can't leak a database error into a response."
>
> "Where does data get translated: wire format to internal types, internal types to storage? Those are the boundaries, and they're where bugs and leaks happen."
>
> "What do I need to test without the rest? If testing a business rule needs a real database and a real HTTP server, the design is telling me something."
>
> "How do errors cross the layers? I want one place that decides what the outside world sees."
>
> "Is this proportionate? Three reasons to change means three layers, not seven. A 50-line script needs no layers at all."

**✍️ Fill in: reasons to change → layers**

| Reason to change | Examples | Layer / package |
|---|---|---|
| ✍️ *Transport concerns* | *URLs, status codes, JSON names, auth headers* | *handlers* |
| ✍️ *Business rules* | *"Unknown email gets defaults", "revision must match"* | *services* |
| ✍️ *Storage concerns* | *Queries, indexes, migrations* | *repositories* |

**✍️ Fill in: what each layer knows and must not know**

| Layer | Knows | Must NOT know |
|---|---|---|
| ✍️ | | |
| ✍️ | | |
| ✍️ | | |

**✍️ Fill in: boundaries**

- Dependency direction: ✍️ *top → down only; consumers define small interfaces for what they need*
- Error flow: ✍️ *storage errors → domain errors (not found / conflict / invalid) → one mapper turns them into external codes; details are logged, never returned*
- Data at the boundary: ✍️ *responses are separate DTOs built by explicit field mapping; internal entities are never serialized directly*
- Package layout: ✍️ *sketch it; accept that it's a two-way door*

**🚩 Junior trap:** using a layered architecture because the tutorial did, then putting logic wherever is convenient. Soon nobody knows where "the" rule lives.

**✅ Exit check**
- [ ] I can name each layer's job and what it must not know.
- [ ] The error flow and DTO boundary are decided.
- [ ] I can explain why there are N layers and not more or fewer.

[Deeper: playbook Phase D](SENIOR-PLAYBOOK.md#phase-d-design-the-shape-architecture-thinking)

---

### A5. Contracts first

**Goal:** define the data model, the API and any stored formats precisely enough that implementing them is mostly mechanical.

**🧠 Inner monologue**
> "A contract is anything something else depends on: the schema, the API, a file format. Once a client depends on it, it's hard to change, so I design it on paper first."
>
> "I'll design the schema around what the server *does* with the data, not what the user *sees*. If the server never reads a field, does it need its own column?"
>
> "Every field gets a reason. If I can't say why a field exists, maybe it shouldn't."
>
> "Are my IDs safe to expose? Sequential integers leak counts and invite guessing."
>
> "For every endpoint: how does the server know who's calling, and what are they allowed to touch? Can I design the URL so the authorization bug has nowhere to live?"
>
> "What does an error look like? One shape everywhere, with stable codes that clients can switch on."
>
> "The riskiest flow gets drawn as a sequence diagram, completely. Drawing it shows the gaps, like 'what happens when X is unknown?'. Every arrow becomes a test."

**✍️ Fill in: data model**

| Entity.field | Type | Why it exists (which requirement or risk) |
|---|---|---|
| ✍️ *users.id* | *UUID* | *Sequential IDs leak user count (R6)* |
| ✍️ *users.email* | *unique, lowercased* | *"A@x" and "a@x" must be one account* |
| ✍️ | | |

Uniqueness, foreign key and delete rules: ✍️ *cascade on user delete; soft delete for items*

How the schema evolves: ✍️ *versioned migration files, reviewed like code*

**✍️ Fill in: API / interface surface**

| Method + path (or function) | Who can call | Ownership enforced by | Notes |
|---|---|---|---|
| ✍️ *GET /api/v1/me* | *Authenticated user* | *ID comes from the token, not the URL* | |
| ✍️ | | | |

Error shape: ✍️ `{"code": "stable_machine_code", "message": "human text"}`

Status / error code table (decide once, never improvise):

| Code | Meaning in this project |
|---|---|
| ✍️ 400 | *Malformed request* |
| ✍️ 401 | *Not authenticated* |
| ✍️ 404 | *Not found, or not yours (don't reveal it exists)* |
| ✍️ 409 | *Conflict (duplicate, stale revision)* |
| ✍️ | |

Versioning strategy: ✍️ */api/v1 prefix; additive changes only within a version*

**✍️ Fill in: the riskiest flow as a sequence diagram**

```
✍️ Client                                   Server
    │── request ──────────────────────────►│
    │                     what's checked?   │
    │                     what if unknown?  │
    │◄──────────────────────── response ────│
```

Write one endpoint spec per endpoint before implementing it (template: [Appendix 1](#appendix-1-mini-templates)).

**🚩 Junior trap:** writing handlers and seeing what URLs come out. The API's shape, and its security properties, end up being accidents of the implementation.

**✅ Exit check**
- [ ] Every endpoint is listed with auth, request, response and errors.
- [ ] Every schema field has a reason.
- [ ] The riskiest flow is drawn, and I can explain it without notes.

[Deeper: playbook Phase E](SENIOR-PLAYBOOK.md#phase-e-design-the-contracts-before-the-code)

---

### A6. Ready to code?

**Goal:** know when design has done its job, kill technical unknowns with spikes, and build a running skeleton before any features.

**🧠 Inner monologue**
> "Design can turn into procrastination. Are the remaining unknowns cheaper to answer in code than on paper? If yes, I start."
>
> "If I'm still debating something fundamental, it's too early. If I'm debating folder names, it's too late. I'm bikeshedding."
>
> "What don't I know about my tools? 'How slow is this?', 'Does this driver actually enforce that constraint?' A one-hour throwaway experiment answers those. A document doesn't."
>
> "Before any feature, I want the boring plumbing real: config, logging, DB, one endpoint, one test, CI. Every feature after that inherits it for free."
>
> "Can I name the first five PRs? If not, I don't understand the work well enough yet."

**✍️ Fill in: gate**
- [ ] I can explain the core flows without notes.
- [ ] All one-way doors are decided and written down.
- [ ] The remaining questions are about tool behavior, not design.

**✍️ Fill in: spikes** (time-boxed, thrown away afterwards; write down what you learned)

| Unknown | Spike (≤ 2 h) | What I learned |
|---|---|---|
| ✍️ *How slow is the hash function with these params?* | *20-line timing program* | |
| ✍️ | | |

**✍️ Fill in: walking skeleton checklist**
- [ ] Starts from validated config (env vars)
- [ ] Structured logging, with request IDs
- [ ] DB connection plus a migration tool with one migration
- [ ] Graceful shutdown and server timeouts
- [ ] One trivial endpoint (e.g. a health check)
- [ ] One test that calls it
- [ ] CI runs lint, tests and build on every PR
- [ ] A `make run` / `make test` (or equivalent) entry point
- [ ] A README that explains how to run it

**✍️ Fill in: first five PRs**
1. ✍️ *init module, entry point, Makefile → health check works*
2. ✍️
3. ✍️
4. ✍️
5. ✍️

**🚩 Junior trap:** starting with the most visible feature and adding plumbing "when we need it", which means retrofitting it under pressure into every feature already built.

**✅ Exit check**
- [ ] The spikes are done and deleted, and what I learned is written down.
- [ ] The skeleton runs locally and in CI.
- [ ] The first feature PR can focus entirely on the feature.

[Deeper: playbook Phase F](SENIOR-PLAYBOOK.md#phase-f-when-to-start-writing-code-and-how-to-start)

---

## Track B: Adding a feature

Use this for every feature in a new project, and for any change to an existing codebase. In an existing codebase, B2 and B3 are the steps juniors skip and seniors never do.

### B1. Clarify the ask

**Goal:** understand the real problem and what "done" means before deciding how to build it.

**🧠 Inner monologue**
> "What was asked is often a solution. What's the problem behind it? 'Add a CSV export' might really mean 'I need to get my data into a spreadsheet once a month'."
>
> "How will anyone know this is done? If I can't write acceptance criteria, I don't understand the ask yet."
>
> "What's the smallest version that solves the real problem? What's tempting to add but isn't needed?"
>
> "Is anything ambiguous enough that I should ask instead of guess? A five-minute question beats a day spent building the wrong thing."

**✍️ Fill in**
- The request as stated: ✍️
- The real problem behind it: ✍️
- Acceptance criteria (observable, testable):
  - [ ] ✍️ *Given X, when Y, then Z*
  - [ ] ✍️
- Out of scope for this change: ✍️
- Open questions to ask before starting: ✍️

**🚩 Junior trap:** building exactly what was literally asked, including the parts that don't solve the actual problem.

**✅ Exit check**
- [ ] I can state the problem in one sentence, without naming a solution.
- [ ] Acceptance criteria are written and testable.

---

### B2. Read before you write

**Goal:** understand how the existing code already solves similar problems, so the new code fits in and reuses what exists.

**🧠 Inner monologue**
> "The codebase already has opinions: how errors are handled, how things are named, where validation lives, how tests are written. I'll match them unless I have a real reason not to, and if I do, I'll say so in the PR."
>
> "Has someone already solved part of this? A helper, a utility, a similar endpoint I can copy the shape of? Writing a second version of something that exists is a maintenance bug."
>
> "Who calls the code I'm about to change? What depends on its current behavior, including behavior nobody intended?"
>
> "What do the existing tests tell me about the intended behavior? And where are there no tests, meaning I need to be careful?"
>
> "Is there a design doc, ADR or roadmap entry about this area? Someone may already have decided something I'm about to re-decide."

**✍️ Fill in**
- Closest existing example to copy the shape of: ✍️ *path/to/similar_feature*
- Existing helpers and utilities to reuse: ✍️
- Conventions to follow (naming, errors, validation, tests): ✍️
- Callers / dependents of the code I'll touch: ✍️
- Relevant docs / ADRs / prior decisions: ✍️
- Areas with weak or no test coverage: ✍️

**🚩 Junior trap:** starting a new pattern in the corner of the codebase you're working in, because you didn't look at how the rest of the code does it.

**✅ Exit check**
- [ ] I've read the code paths I'm about to change, end to end.
- [ ] I know which existing pieces I'll reuse.

---

### B3. Blast radius and compatibility

**Goal:** know what this change can break, and how to undo it if it does.

**🧠 Inner monologue**
> "What existing behavior changes? Not just what I intend to change: what else moves with it?"
>
> "Does this touch a contract? The API, the schema, a config key, a file format, an event? Who depends on it, and will old clients still work?"
>
> "Is there existing data? Does it need a migration or a backfill? What does old data look like when new code reads it?"
>
> "Can this ship in steps? Add the new thing first, migrate, and only then remove the old one (expand, then contract)."
>
> "If this goes wrong after merging, what's the rollback? Revert the commit? Is the migration reversible? Do I need a feature flag?"
>
> "What's the worst realistic failure, and how would I even notice?"

**✍️ Fill in**
- Behavior that changes (intended): ✍️
- Behavior that might change (unintended): ✍️
- Contracts touched (API / schema / config / format): ✍️
- Backwards compatible? ✍️ *Yes / No → plan:*
- Data migration or backfill needed: ✍️
- Rollout plan: ✍️ *single PR / expand-then-contract / behind a flag*
- Rollback plan: ✍️
- How I'd notice it broke: ✍️ *failing test, error-rate metric, log line*

**🚩 Junior trap:** a migration that works on an empty dev database but fails, or loses data, on real existing rows.

**✅ Exit check**
- [ ] Every touched contract is listed, with a compatibility answer.
- [ ] There's a rollback plan I'd actually be able to follow.

---

### B4. Risk and one-way doors

**Goal:** find the part of this feature most likely to go wrong, and the decisions that are hard to reverse.

**🧠 Inner monologue**
> "What part of this am I most nervous about? That's where I start, and where most of the tests go."
>
> "Which decisions in this feature are one-way doors? A new public field, a stored format, a new table. Those get written down before I code."
>
> "Does this touch anything from the risk table (A2)? Auth, other users' data, money, deletes, concurrency?"
>
> "Does this add a defense that creates a new attack? A lockout that lets anyone lock out a real user, for example."

**✍️ Fill in**
- Riskiest part: ✍️
- One-way doors in this feature (and where they're written down): ✍️
- Risk-table entries touched: ✍️ *R2 (cross-user access)*
- New risks introduced: ✍️

**🚩 Junior trap:** starting with the easy, familiar part of the feature and leaving the risky part for the end, when there's least time to get it right.

**✅ Exit check**
- [ ] The riskiest part is identified and will be built first.
- [ ] One-way doors are decided on paper (ADR or design note).

---

### B5. The four questions and the test list

**Goal:** turn the feature into a function signature and a complete test list before writing the implementation.

**🧠 Inner monologue**
> "Four questions for any piece of code. What goes in, and which of it is untrusted? What comes out? What can go wrong? What must never happen?"
>
> "The answers to 1 and 2 are the signature. The answers to 3 and 4 are the test list. Once I have both, the body is usually the easy part."
>
> "What if two of these run at the same time? Any 'check, then act' is a race unless something like a unique index or a transaction makes it atomic."
>
> "This list is my definition of done. If it's not on the list, I'm not done thinking. If it is on the list, it gets a test."

**✍️ Fill in**

1. **What goes in?** (mark untrusted inputs) ✍️ *email (untrusted), params (untrusted), caller ID (trusted, from token)*
2. **What comes out?** ✍️ *created resource ID or a domain error*
3. **What can go wrong?** ✍️ *malformed input, not found, duplicate, store error, timeout*
4. **What must never happen?** ✍️ *secret returned in the response, another user's data touched, silent overwrite, internal error text reaching the client*
5. **What if two run at once?** ✍️ *both pass the uniqueness check → rely on a DB unique constraint instead*

**Test list** (weight by risk: the "must nevers" get the most tests)

```
✍️ core rules (unit, with fakes):
  [ ] happy path → expected result
  [ ] each "can go wrong" → the right domain error
  [ ] each "must never" → asserted explicitly

✍️ boundary (transport / integration):
  [ ] valid request → right status and body, no sensitive fields
  [ ] malformed request → 400 with a stable code
  [ ] another user's resource → 404

✍️ storage (against the real DB):
  [ ] constraints actually enforced (unique, FK)
```

**🚩 Junior trap:** writing code until it seems to work, then wondering how to test it. The tests end up checking what the code *does*, not what it *should* do.

**✅ Exit check**
- [ ] The signature is decided.
- [ ] Every "can go wrong" and "must never" has a matching test on the list.

[Deeper: playbook Phase G](SENIOR-PLAYBOOK.md#phase-g-how-a-senior-knows-what-code-to-write)

---

### B6. Slice into PRs

**Goal:** split the feature into small PRs that are each correct, mergeable and quick to review.

**🧠 Inner monologue**
> "This feature isn't one PR. What's the thinnest slice that delivers one working, tested behavior?"
>
> "Each slice must leave `main` working. I never merge half a feature that breaks something."
>
> "Can someone review this in about 15 minutes? Big PRs get worse reviews, not better ones."
>
> "Does each PR have exactly one reason to exist? Refactors, renames and dependency upgrades go in their own PRs."
>
> "In what order do I build each slice? From the contract inward, but with the core rules first, because that's where the dangerous mistakes are."

**✍️ Fill in: slice plan**

| # | PR (one behavior) | Depends on | Leaves `main` working? |
|---|---|---|---|
| 1 | ✍️ *migration + repository + repository tests* | — | ✔ |
| 2 | ✍️ *service rule + unit tests* | 1 | ✔ |
| 3 | ✍️ *endpoint + handler tests + wiring* | 2 | ✔ |

**Build order inside each slice (outside-in, rules first)**

```
1. Test list (from B5)
2. Signature + the small interface the logic needs
3. Unit tests for the core rules, with fakes   ← the rules live here
4. Implement until they pass
5. Storage: migration + repository + test against the real DB
6. Transport: handler + test
7. Wire it up (main / router / DI)
8. Run it manually once, end to end
```

**🚩 Junior trap:** building horizontally (all models, then all repositories, then all handlers), so nothing works end to end until the very end.

**✅ Exit check**
- [ ] Every slice has one reason to exist and leaves `main` green.
- [ ] I know what PR #1 is.

---

### B7. The build loop

**Goal:** always know what to write next, and notice when reality disagrees with the design.

**🧠 Inner monologue**
> "What do I write next? The next failing test on the list. If there isn't one, either I'm done or the list is incomplete."
>
> "This function is hard to test. That's design feedback: it's probably doing too much, or depends on something concrete it shouldn't."
>
> "I just discovered something that contradicts the design. I stop, update the design doc or ADR, and then continue. The code shouldn't silently drift away from the paper."
>
> "I'm tempted to fix something unrelated. I'll note it and do it in a separate PR."
>
> "I've been stuck for longer than my time box. Time to step back, re-read the four questions, or ask someone."

**✍️ Fill in (running log while building)**
- Design changes discovered (and where I recorded them): ✍️
- Unrelated things I noticed (for later PRs): ✍️
- Questions to raise in review: ✍️

**🚩 Junior trap:** "while I'm here" scope creep, which turns a 100-line PR into a 1,000-line one that nobody can review properly.

**✅ Exit check**
- [ ] Every test on the list passes.
- [ ] The design doc matches what was actually built.

---

### B8. Self-review

**Goal:** catch your own mistakes before anyone else has to.

**🧠 Inner monologue**
> "I'll read my diff as if a stranger wrote it and I'm looking for the bug."
>
> "First correctness, then security boundaries, then design, then tests, then readability. Style last, and only what the linter missed."
>
> "Would I understand this in six months, with no context?"

**✍️ Fill in: self-review checklist**
- [ ] **Correctness:** it does what the PR says. I've thought about which inputs break it.
- [ ] **Boundaries:** untrusted input is validated. Every query is scoped to the caller.
- [ ] **Leaks:** no response or log contains secrets, tokens, internal error text or personal data it shouldn't.
- [ ] **Errors:** every error is handled or deliberately returned, with its cause. Nothing is silently ignored.
- [ ] **Design:** logic sits in the right layer and follows existing patterns (B2).
- [ ] **Tests:** they cover failures and "must nevers", not just the happy path.
- [ ] **Enforcement:** no "remember to…" rule that the code could enforce instead.
- [ ] **Cleanliness:** no dead code, no commented-out code, no debugging leftovers.
- [ ] **Comments:** they explain *why*, not what the signature already says.
- [ ] **Scope:** the PR has one reason to exist.

**🚩 Junior trap:** opening the PR without reading your own diff. You'd catch about a third of the issues yourself.

**✅ Exit check**
- [ ] Every box above is ticked, or the exception is explained in the PR description.

[Deeper: playbook Phase I](SENIOR-PLAYBOOK.md#phase-i-quality-as-a-system-not-an-afterthought)

---

### B9. Ship and verify

**Goal:** confirm the change actually works where it matters, and leave a record of why it was made.

**🧠 Inner monologue**
> "'It should work' isn't evidence. Did CI actually run and go green? Did I run it myself?"
>
> "The diff shows *what*. The PR description has to explain *why*, because future me will read it."
>
> "Did this change any decisions or docs? The README, the API spec, an ADR, the roadmap?"
>
> "After merging: how will I know it's healthy? Which log line or metric would show a problem?"

**✍️ Fill in**
- [ ] CI is green (I checked the actual run, not just assumed)
- [ ] Ran it end to end manually: ✍️ *what I did, what I saw*
- [ ] Acceptance criteria from B1 are all met
- [ ] Docs / ADRs / API spec updated: ✍️
- [ ] PR description written (template: [Appendix 1](#appendix-1-mini-templates))
- [ ] After merge, checked: ✍️ *logs, metrics, health check*

**🚩 Junior trap:** assuming instead of verifying ("CI is set up", but the workflow never actually ran).

**✅ Exit check**
- [ ] I have evidence, not a feeling, that it works.

---

## Track C: Done and operating

**Goal:** be able to run the system, notice when it's unhealthy, recover from the bad day, and finish deliberately.

**🧠 Inner monologue**
> "How do I know it's working right now? How would I find out it's broken before a user tells me?"
>
> "What's the worst realistic bad day, and what do I do on it? A backup I've never restored is a hope, not a backup."
>
> "Could someone else run this from the docs alone? Could I, in a year?"
>
> "Does the system actually keep the core promise from A1? Can I prove it with a test?"
>
> "What did I knowingly leave out? Writing limitations down isn't admitting failure. It shows I know where the edges are."

**✍️ Fill in: bad-day table**

| Bad day | My answer | Tested? |
|---|---|---|
| ✍️ *Disk dies* | *Backups + restore procedure* | *Restored once on (date)* |
| ✍️ *A bad migration ships* | | |
| ✍️ *A secret / key leaks* | | |
| ✍️ *A dependency has a CVE* | | |
| ✍️ *It's slow or erroring* | | |

**✍️ Fill in: observability**
- Health check(s): ✍️ *liveness / readiness*
- Key metrics: ✍️ *request rate, error rate, latency*
- Logs: ✍️ *structured, with request IDs, never containing secrets*
- Runbook location: ✍️ *docs/runbook.md: start, stop, upgrade, backup, restore, rotate secrets*

**✍️ Fill in: definition of done**
- [ ] Each promise from A1 is verified with evidence: ✍️
- [ ] Every "must never" has a test
- [ ] A restore from backup has been done at least once

**Known limitations (deliberate)**
- ✍️ *Limitation, and why it's accepted*

**Later list** (thought about, deliberately not in this version)
- ✍️ *Feature: what it would require, in one line*

**🚩 Junior trap:** never finishing ("just one more feature"), or declaring victory without checking the original goals.

[Deeper: playbook Phases J-K](SENIOR-PLAYBOOK.md#phase-j-operate-it)

---

## Appendix 1: Mini templates

### ADR (Architecture Decision Record)

```markdown
# ADR-NNN: <decision as a short imperative phrase>

Date: YYYY-MM-DD
Status: Proposed | Accepted | Superseded by ADR-MMM

## Context
What forces a decision? Which constraints apply? (Facts, not opinions.)

## Decision
What we're doing, in one or two sentences.

## Alternatives considered
- Option A: why not
- Option B: why not

## Consequences
- + good things that follow
- − costs we accept

## Revisit when
The concrete condition that would make us reconsider.
```

Keep ADRs numbered in `docs/adr/`. Never edit an accepted one. Supersede it with a new one, so the history of *why* is preserved.

### Design doc

```markdown
# <Feature> design

## Goal            One paragraph: what problem, for whom.
## Non-goals       What this explicitly doesn't do.
## Background      What the reader needs first. Links to ADRs / risk table.
## Design          How it works. Diagrams. Data model and API changes.
## Risks           Which risks it touches. What must never happen.
## Compatibility   Contracts touched, migration, rollout, rollback.
## Alternatives    And why they lost.
## Testing plan    The test list.
## Open questions  What's still undecided.
```

Two to five pages. If it's longer, the feature is probably too big. Split it.

### Endpoint / interface spec

```markdown
## METHOD /path   (or: func Name(...))

Purpose:
Auth / caller:            Rate limit:
Input:     fields, types, validation rules (mark untrusted)
Output:    status + body (or return type)
Errors:    status + code, and when each happens
Security notes:  enumeration? ownership? what must not be logged?
Tests:     the list
```

### PR description

```markdown
## What
One or two sentences.

## Why
The reason, with a link to the roadmap step / design doc / issue.

## How
Anything non-obvious about the approach. Alternatives you rejected.

## Testing
Tests added; what you checked manually.

## Notes for the reviewer
Where you're unsure; what to look at first.
```

### Risk table row

```markdown
| # | Threat | What it can do | Worst outcome | Mitigation | Residual risk (and why accepted) |
|---|--------|----------------|---------------|------------|----------------------------------|
```

---

## Appendix 2: When I'm stuck

| Stuck on… | Ask yourself |
|---|---|
| **What to build at all** | What's the core promise? Who would stop using this, and why? (A1) |
| **What to build next** | What am I most nervous about? What does everything else depend on? |
| **What code to write next** | What's the next failing test on my list? If there's none, are the four questions answered? (B5) |
| **How to structure it** | What are the reasons this will change? What must this part *not* know? (A4) |
| **A decision I can't make** | Is it a one-way or a two-way door? If two-way: pick one in 60 seconds and move on. |
| **A design that feels wrong** | Is it hard to test? Does it rely on someone remembering a rule? |
| **Naming** | What would a reader with no context expect this to do? Use that. It's a two-way door anyway. |
| **A surprising review comment** | Find it in Appendix 3. Most review comments are one of those patterns. |
| **A "can you also add…" request** | Does it serve the real problem from B1? If not, put it on the later list and a separate PR. |
| **Too long on one thing** | Was it time-boxed? Overrunning is a finding in itself: step back, spike it, or ask. |

---

## Appendix 3: Anti-patterns to catch in yourself

| # | Anti-pattern | Looks like | Senior alternative |
|---|---|---|---|
| 1 | Starting from the tutorial shape | Generic CRUD on day one | Start from the core promise and the risks |
| 2 | Quality / security as a later feature | "We'll add it in sprint 4" | Decide one-way doors before the schema and API |
| 3 | Trusting IDs from the caller | `PUT /things/:id` with no ownership check | Identity from the token; scope every query |
| 4 | Returning internal entities | Serializing the DB model directly | Explicit DTOs at the boundary |
| 5 | Raw errors to clients | `"UNIQUE constraint failed: …"` | Domain errors → stable codes; log the details |
| 6 | Check-then-act | `if !exists { insert }` | A DB constraint or transaction makes it atomic |
| 7 | Silent overwrite | Last write wins | Revision checks + a conflict error |
| 8 | Plumbing last | Config, logging, CI added after ten features | Walking skeleton first |
| 9 | Huge PRs | "Implement auth" in 2,000 lines | One slice per PR, ~15 minutes to review |
| 10 | Happy-path-only tests | One test: "it works" | Test list from "can go wrong" and "must never" |
| 11 | Untestable design | Business logic holding a concrete DB handle | Small consumer-defined interfaces + fakes |
| 12 | Over-engineering | Microservices for a solo app | The simplest thing that meets the NFRs |
| 13 | Under-engineering one-way doors | "Auto-increment IDs for now" | Decide before the first migration |
| 14 | Ignoring the existing codebase | A new pattern for a solved problem | Read first (B2); match conventions |
| 15 | Ignoring blast radius | A migration tested only on an empty DB | B3: compatibility, data, rollback |
| 16 | Assuming instead of verifying | "CI is set up" (it never ran) | Check the green tick; restore the backup |
| 17 | Comments that repeat the code | `// GetUser gets user` | Explain what the signature doesn't show |
| 18 | Keeping dead code "just in case" | Commented-out blocks | Delete it; git remembers |
| 19 | Logging secrets | Logging request bodies on auth routes | Structured fields only; never secrets |
| 20 | Unbounded input | Accepting any body size | Size and count limits at the boundary |

[Deeper: playbook Appendix C](SENIOR-PLAYBOOK.md#appendix-c-junior--senior-anti-patterns)

---

*The template is scaffolding. After a few projects you'll stop needing to fill it in, because you'll ask "what must never happen here?" and "is this reversible?" without thinking about it. That habit is the actual skill.*
