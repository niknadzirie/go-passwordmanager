# The Senior Engineer's Playbook: Building a Zero-Knowledge Password Manager from Scratch

> **What this is:** a walkthrough of *how a senior engineer thinks* when given this project with an empty repository. It covers the questions they ask, the order they make decisions in, how they choose an architecture, how they know what code to write next, and when they actually start typing.
>
> **What this is not:** a task list. [`ROADMAP.md`](ROADMAP.md) is the task list for *this* repo, starting from where the code is today. This playbook explains the reasoning *behind* a roadmap like that, so that next time you can write one yourself.

---

## Table of contents

- [Part 0: How to read this](#part-0-how-to-read-this)
- [Part 1: The senior mindset](#part-1-the-senior-mindset)
- [Phase A: Understand the problem](#phase-a-understand-the-problem-no-code)
- [Phase B: Find the one thing that must not go wrong](#phase-b-find-the-one-thing-that-must-not-go-wrong)
- [Phase C: Constraints and choosing technology](#phase-c-constraints-non-functional-requirements-and-choosing-technology)
- [Phase D: Design the shape (architecture)](#phase-d-design-the-shape-architecture-thinking)
- [Phase E: Design the contracts before the code](#phase-e-design-the-contracts-before-the-code)
- [Phase F: When to start writing code, and how](#phase-f-when-to-start-writing-code-and-how-to-start)
- [Phase G: How a senior knows what code to write](#phase-g-how-a-senior-knows-what-code-to-write)
- [Phase H: Build features in risk order](#phase-h-build-features-in-risk-order)
- [Phase I: Quality as a system](#phase-i-quality-as-a-system-not-an-afterthought)
- [Phase J: Operate it](#phase-j-operate-it)
- [Phase K: Know when you're done](#phase-k-know-when-youre-done)
- [Appendix A: Question bank](#appendix-a-question-bank)
- [Appendix B: Templates](#appendix-b-templates)
- [Appendix C: Junior → senior anti-patterns](#appendix-c-junior--senior-anti-patterns)
- [Appendix D: Mapping to ROADMAP and this repo](#appendix-d-mapping-to-roadmap-and-this-repo)

---

## Part 0: How to read this

The single biggest difference between a junior and a senior engineer is not how fast they type or how many libraries they know. It's **where they spend their thinking**.

- A junior spends most of their thinking *inside* the code: "how do I make this work?"
- A senior spends a large share of their thinking *before* and *around* the code: "what should exist at all, in what order, what could go wrong, and how will I know?"

So most of this document happens before the first `go mod init`. That is deliberate. When a senior does start writing code, they write it quickly and with little rework, because the hard questions have already been answered.

Each phase below has the same structure:

1. **The goal** of the phase, in one sentence.
2. **The questions** a senior asks. Learn these; they are the actual skill.
3. **The reasoning**, i.e. how the answers turn into decisions.
4. **A worked example** using this project.
5. **A junior vs senior callout** that makes the difference in thinking concrete.
6. **Exit criteria**: how a senior knows the phase is done.

The phases are presented in order, but real work loops back. You'll discover something in Phase G that changes a decision from Phase E. That's normal. The skill is to notice it, go back and update the decision *on paper*, and then continue. Don't let the code silently drift away from the design.

> **A note on code snippets:** snippets here are short illustrations of an idea, not implementations to copy. You still write the code (see the working agreement in `CLAUDE.md`).

---

## Part 1: The senior mindset

These principles show up in every phase. Read them first, then watch for them later.

### 1.1 Order work by risk, not by ease

Juniors naturally start with what they know how to do: a CRUD endpoint, a model, a nice folder structure. It feels productive. But the easy parts were never going to sink the project.

Seniors ask: **"What is most likely to make this project fail, or force a rewrite?"** and do *that* first. For a password manager, the answers are:

1. Getting the security model wrong (it changes everything: schema, API, auth).
2. Getting authentication or authorization wrong (a data breach).
3. Losing user data (sync conflicts, bad deletes, no backups).

Notice that "writing a GET endpoint" isn't on the list. It's easy and low risk, so it can wait.

> **Rule of thumb:** if you're unsure what to do next, pick the thing you're most *nervous* about. Nervousness is a good risk detector.

### 1.2 One-way doors and two-way doors

Some decisions are cheap to reverse (a "two-way door"): the name of a function, which logging library you use, folder structure inside `internal/`. Others are expensive or impossible to reverse (a "one-way door"): the encryption scheme once users have data, the public API once clients depend on it, primary key types once data exists.

Seniors **spend design time in proportion to how irreversible a decision is**:

| Decision | Door type | How much thought |
|---|---|---|
| Zero-knowledge vs server-side encryption | One-way | Days. Write it down. Get it reviewed. |
| Public API shape (`/me` vs `/users/:id`) | Mostly one-way | Hours. Write a spec. |
| Primary key type (UUID vs int) | Expensive to change later | An hour. Decide before the first migration. |
| Gin vs chi vs stdlib router | Two-way (behind handlers) | Minutes. Pick one and move on. |
| Variable names, file names | Two-way | Seconds. Rename later. |

The junior mistake runs in both directions: agonizing over two-way doors (which router? which folder name?) while rushing through one-way doors (just store the password, we'll encrypt it later).

### 1.3 Walking skeleton first, then vertical slices

A **walking skeleton** is the thinnest possible end-to-end version of the system that actually runs: a binary that starts from config, connects to the DB, serves one trivial endpoint, logs properly, shuts down cleanly, and passes CI. It does almost nothing, but *all the plumbing is real*.

After that, every feature is a **vertical slice**: a thin cut through every layer (route → handler → service → repository → migration → test) that delivers one working behavior.

```
 Horizontal (junior)                    Vertical (senior)
 ┌──────────────────────────┐           ┌────┬────┬────┬────┐
 │ all models               │           │    │    │    │    │
 ├──────────────────────────┤           │ R  │ L  │ R  │ V  │
 │ all repositories         │           │ e  │ o  │ e  │ a  │
 ├──────────────────────────┤           │ g  │ g  │ f  │ u  │
 │ all services             │           │ i  │ i  │ r  │ l  │
 ├──────────────────────────┤           │ s  │ n  │ e  │ t  │
 │ all handlers             │           │ t  │    │ s  │    │
 ├──────────────────────────┤           │ e  │    │ h  │ .. │
 │ "now let's see if it     │           │ r  │    │    │    │
 │  works end to end"       │           │ ✔  │ ✔  │ ✔  │    │
 └──────────────────────────┘           └────┴────┴────┴────┘
  Nothing works until the end.           Something works after every PR.
```

*Why:* vertical slices give you feedback early. If the design is wrong, you find out after one slice, not after building every layer.

### 1.4 Make it work → make it right → make it fast

Kent Beck's ordering. In ordinary code, "make it work" can be a bit scrappy. In **security code, "work" includes "right"**: a login endpoint that accepts the right password but also leaks whether an email exists does not "work". Seniors adjust the definition of "work" to the domain.

"Make it fast" almost never applies to a personal password manager, with one important exception: things that are *deliberately* slow (Argon2id) need their cost tuned. That's a security parameter, not a performance optimization.

### 1.5 Write things down

Seniors write short documents, and not for bureaucracy. **Writing is thinking.** You can't hold a vague idea in a document; the act of writing forces you to notice the gaps.

The three kinds of writing used throughout this playbook:

- **Problem statement** (1 page): what are we building, for whom, and what's out of scope.
- **Design doc** (2-5 pages): how it works, with alternatives considered.
- **ADR** (Architecture Decision Record, half a page): one decision, its context, and its consequences. These form a log you can read later to understand *why* the code is the way it is.

### 1.6 Choose boring technology; minimize dependencies

Every dependency is code you didn't write, can't fully review, and must keep updating. Every new technology is something new that can break in ways you don't understand yet.

Seniors ask of each dependency: **"What does this save me, and what does it cost me?"** A JWT library saves you from writing a security-sensitive parser: worth it. A library that wraps `strings.ToLower`: not worth it.

For a password manager specifically, the crypto rule is absolute: **never invent primitives; compose standard ones exactly as specified**, from the standard library or `golang.org/x/crypto`.

### 1.7 Optimize for the reader and for change

Code is read far more often than it's written, and it's read most by "future you" who has forgotten everything. Seniors ask:

- "Will this be obvious to me in six months?"
- "If requirement X changes, how many files do I touch?"
- "Can I delete this later without fear?"

Cleverness is a cost. A slightly longer, boring, explicit function beats a short clever one. (You saw this in the `phase0/ci` review: the explicit DTO mapping is "longer" than the struct conversion, but it's the version that stays correct when `User` grows.)

### 1.8 Make the right thing easy and the wrong thing hard

The best senior designs don't rely on everyone remembering a rule. They make the rule impossible to break by accident.

- Instead of "remember to check the user ID matches", design the API so there *is* no user ID in the URL (`/me`).
- Instead of "remember not to return the password hash", make the response type a DTO that has no such field.
- Instead of "remember to run the linter", make CI block merges until it passes.

When you catch yourself writing "remember to…" in a doc, ask whether the design could enforce it instead.

---

## Phase A: Understand the problem (no code)

**Goal:** be able to explain, in one page, what is being built, for whom, what success looks like, and what you're *not* building.

### The questions a senior asks

**Who and why**
- Who uses this? (One person? A family? A company?)
- What problem do they have today? (Reusing passwords. Sticky notes. Browser storage they don't trust.)
- Why would they trust *this* product with their most sensitive data?

**The core promise**
- What is the one sentence this product must make true? For a password manager: *"Only you can read your passwords, even if our servers are compromised."*
- What would make a user stop using it immediately? (A breach. Losing their data. Being locked out.)

**Scope**
- What's the minimum that's useful? (Store, retrieve and sync login credentials, securely.)
- What's explicitly out of scope *for now*? (Sharing, organizations, browser extension, autofill, attachments, passkeys.)
- What's out of scope *forever*? (Recovering a forgotten master password, which is impossible by design in zero-knowledge.)

**Context**
- Who builds it, and how much time do they have? (One person, 1-2 hours a day.)
- Where will it run? (Locally, Docker Compose.)
- What's the real purpose? (A learning/portfolio project. That matters: it makes "explain your trade-offs" a deliverable in itself.)

### The reasoning

The scope questions are where seniors earn their keep. Every feature you say "no" to now is weeks you don't spend later, and complexity you don't have to secure.

Look at what saying "no sharing" did for this project. Sharing in a zero-knowledge system needs public/private key pairs per user, organization keys, key distribution, and revocation (what happens when someone leaves an org?). Saying no removed an entire category of crypto and an entire category of authorization bugs.

Equally important is noticing **consequences of the core promise**. "Only you can read your passwords" *implies* "if you forget your master password, your data is gone." That's not a bug to fix later; it's a product decision to communicate clearly. A senior surfaces implications like this in the first hour, before anyone builds a "forgot password" button that can't work.

### Worked example: the problem statement for this project

```markdown
# Problem statement: Vault (working name)

## What
A self-hosted password manager. A Go CLI stores and syncs a user's
credentials through a small REST API. Zero-knowledge: the server never
sees plaintext secrets or the master password.

## For whom
Individuals running it themselves (and portfolio reviewers reading the code).

## Core promise
If the server, database or backups are stolen, the attacker learns nothing
useful about any vault contents.

## Must have (v1)
- Register, log in, log out; change master password
- Create/read/update/delete encrypted items and folders
- Sync between multiple devices of the same user, without silent data loss
- Basic abuse protection on auth (rate limiting)

## Explicitly NOT in v1
- Sharing between users, organizations
- Browser extension / autofill / web UI
- Attachments, passkeys
- Password recovery (impossible by design; we will say so clearly)

## Constraints
- One developer, ~1-2 h/day
- Runs locally via Docker Compose; SQLite
- Learning project: tests, CI, docs and explained trade-offs are deliverables

## Success looks like
- A reviewer can `docker compose up`, use the CLI, and read in the README
  *why* the design is safe.
- A full DB dump contains no readable secrets (verified by a test).
```

That's the whole thing. It took maybe an hour, and every later decision can be checked against it.

> **Junior vs senior**
>
> A **junior** gets "build a password manager", opens the editor, and runs `go mod init`. Within a day there's a `User` model with a `Password` field and a `/users` CRUD API, because that's the shape every tutorial starts with.
>
> A **senior** opens a blank document and asks "what does this product promise?" They realize within the first hour that the promise ("only you can read it") rules out the tutorial shape entirely, *before* writing a line that would need to be thrown away.

### Exit criteria
- One-page problem statement exists.
- You can say what's out of scope without hesitating.
- You've listed the consequences of the core promise (e.g. "no password recovery").

---

## Phase B: Find the one thing that must not go wrong

**Goal:** identify what you're protecting, from whom, and decide the security model, because it shapes everything after it.

### The questions a senior asks

- **What are the assets?** (Vault contents. The master password. Session tokens. The list of who has an account.)
- **Who are the attackers, and what can each one do?**
- **Where are the trust boundaries?** Where does data cross from something you control to something you don't?
- **What happens when (not if) each component is compromised?**
- **Which of these decisions are one-way doors?**

### The reasoning: the question that decides the architecture

There's one question that, for a password manager, decides more than any other:

> **"If someone steals the database, what can they read?"**

There are essentially two answers:

1. **Server-side encryption.** The server encrypts vault items with a key it holds. A stolen DB alone is unreadable, but anyone who controls the server (an attacker with a shell, a malicious operator, a subpoena) can read everything, because the key is right there.
2. **Zero-knowledge (client-side / end-to-end) encryption.** The client derives keys from the master password and encrypts everything before it leaves the device. The server stores only ciphertext and can't decrypt it even if it wants to.

Option 1 is simpler. Option 2 is what the core promise from Phase A *requires*. So the decision is already made by the problem statement, and a senior notices that rather than debating it.

Now notice **why this must be decided first**. It's not a feature; it changes the shape of everything:

| Area | Server-side encryption | Zero-knowledge |
|---|---|---|
| What the client sends at login | The password | A hash derived from the password (never the password) |
| User table | `password_hash` | KDF params, hash-of-auth-hash, *wrapped* user key |
| Vault item table | Columns like `username`, `password` (encrypted) | One opaque `data` blob |
| Server-side validation | Can validate fields | Can only validate size/format |
| Search | Server can search | Client must search locally |
| Password change | Rehash | Re-wrap the user key client-side |
| Forgot password | Possible | Impossible |
| Where crypto code lives | Server | Client |

If you start with option 1's shape (the tutorial shape) and switch later, you rewrite the schema, the API, auth, and migrate every user's data. That's a one-way door, which is why it's Phase B and not Phase H.

### Worked example: a threat model

A threat model doesn't need to be fancy. A table is enough. For each attacker, ask what they can do and what limits the damage.

| # | Attacker | Capability | What they could get | Mitigation | Residual risk |
|---|---|---|---|---|---|
| T1 | DB thief | Copy of `vault.db` or a backup | Ciphertext, KDF params, server-side hashes, emails | Zero-knowledge encryption; Argon2id on the auth hash; strong KDF on the client | Offline brute force of weak master passwords; email list exposed |
| T2 | Server compromise | Code execution on the server | Everything T1 has, plus live tokens and requests | Server never receives the master password or keys; short-lived access tokens | Attacker could serve a malicious client update (out of scope: CLI built from source) |
| T3 | Network attacker | Sees/modifies traffic | Tokens, auth hash | TLS; auth hash is not the encryption key | Misconfigured TLS (CLI refuses plain HTTP) |
| T4 | Online guesser | Calls the login API | Account access via weak passwords | Rate limiting, lockout, 2FA, slow KDF | Distributed attacks (partial mitigation) |
| T5 | Other user | A valid account of their own | Another user's items | Every query scoped by the token's user ID; no IDs from URLs for "self" | Bugs in scoping (tests per endpoint) |
| T6 | Enumerator | Calls public endpoints | Who has an account | Prelogin returns defaults for unknown emails; identical login errors and timing | Registration endpoint still reveals "email taken" (accepted, documented) |
| T7 | Stolen device | The user's laptop | Cached tokens, local encrypted cache | Refresh token rotation, revoke sessions, nothing decrypted at rest | Unlocked session on an unlocked laptop |

Look at the "Residual risk" column. A senior doesn't pretend everything is solved; they **write down what's accepted** and why. "Registration reveals that an email is taken" is a deliberate trade-off (the alternative, email-confirmation-before-anything, adds a lot of complexity). Writing it down turns an oversight into a decision.

### Worked example: the trust boundary diagram

```
   ┌──────────────── TRUSTED (user's device) ─────────────────┐
   │                                                           │
   │   master password ──► KDF ──► master key ──► user key     │
   │                                     │            │        │
   │                                     ▼            ▼        │
   │                               auth hash     encrypt/decrypt│
   │                                     │        vault items   │
   └─────────────────────────────────────┼────────────┼────────┘
                                         │ TLS        │ ciphertext only
   ══════════════════ TRUST BOUNDARY ════╪════════════╪═════════════
                                         ▼            ▼
   ┌──────────────── UNTRUSTED (server, DB, backups) ─────────────┐
   │  stores: Argon2id(auth hash), KDF params, wrapped user key,   │
   │          encrypted blobs, refresh-token HASHES                │
   │  never sees: master password, master key, user key, plaintext │
   └───────────────────────────────────────────────────────────────┘
```

From the server's point of view, *the server itself is untrusted*. That sounds strange, but it's exactly the mindset that makes the design robust: you design the server so that its compromise is survivable.

### How this phase shapes everything after it

Before any code exists, the threat model has already decided:

- **The data model:** no plaintext columns, a generic encrypted blob (Phase E).
- **The API:** `/me` style endpoints (T5), generic errors (T6), prelogin with defaults (T6).
- **The auth design:** short access tokens plus revocable refresh tokens (T2, T7).
- **Operational requirements:** rate limiting (T4), TLS (T3), backups that are safe to lose (T1).

This is the senior move: **requirements come from threats, not from a list of "features most apps have".**

> **Junior vs senior**
>
> A **junior** treats security as a feature: "We'll add encryption in sprint 4." Then they discover that adding it means changing every table, every endpoint and every client.
>
> A **senior** treats security as a *property of the architecture*. They ask "what can an attacker read?" before designing a single table, because the answer *is* the design.

### Exit criteria
- The security model is chosen and written down (`docs/design/security-model.md`).
- A threat model table exists, including accepted residual risks.
- You can draw the trust boundary from memory.

---

## Phase C: Constraints, non-functional requirements, and choosing technology

**Goal:** pick the tools from the constraints, not from habit or hype, and write down why.

### The questions a senior asks

**Non-functional requirements (NFRs)**, i.e. "how well" rather than "what":
- **Scale:** how many users, how much data, how many requests per second? (Here: a handful of users, a few thousand items, single-digit requests per second.)
- **Availability:** what happens if it's down for an hour? (Annoying, not catastrophic. The CLI can use its local cache.)
- **Durability:** what happens if data is lost? (Catastrophic. Backups are a must.)
- **Latency:** does anything need to be fast? (No, except it mustn't feel broken. Login is *deliberately* slow because of the KDF.)
- **Security:** see Phase B. This is the dominant NFR.
- **Operability:** who runs it, and how? (One person, Docker Compose, locally.)

**Team and time constraints:**
- What does the team already know? (Go.)
- How much time is there? (1-2 hours a day: minimize moving parts.)
- Who maintains it? (The same person, forever. So boring and simple wins.)

### The reasoning: deriving the stack from the answers

Watch how each choice falls out of a constraint, rather than being a preference:

| Choice | Constraint it comes from | The trade-off accepted |
|---|---|---|
| **Go** | Owner knows Go; one static binary is easy to deploy; strong stdlib crypto (`crypto/*`, `x/crypto`) | More verbose than some languages |
| **REST + JSON** | One simple client (CLI); easy to debug with curl | Less efficient than gRPC; irrelevant at this scale |
| **Gin** | Popular, well-documented, adequate | A dependency; stdlib `net/http` (Go 1.22+ routing) would also work. Two-way door, so don't agonize |
| **GORM** | Familiar, reduces boilerplate | Hides SQL; you must still understand the queries (and use migrations, not AutoMigrate, for prod) |
| **SQLite** | Tiny scale, one machine, one operator, durability via simple backups | No horizontal scaling; one writer at a time |
| **Go CLI client** | Crypto must run on the client; reuse Go; no browser means no CORS/XSS surface | Less friendly than a GUI |
| **Docker Compose** | Local deploy; reviewers can run it with one command | Not a production orchestrator (and doesn't need to be) |
| **Mailpit** | Email features without a real provider | Behind a `Mailer` interface, so a real provider is a config change |

Notice what *isn't* chosen: Kubernetes, microservices, Redis, a message queue, Postgres. Each would be reasonable in some other project. Here, each would add operational work and failure modes without addressing any constraint. **The best architecture is the simplest one that meets the requirements.**

### "But will it scale?"

Juniors sometimes choose Postgres or microservices "in case it needs to scale". A senior answers this with numbers:

- SQLite in WAL mode on an ordinary SSD handles thousands of reads per second and hundreds of writes per second.
- A very heavy password manager user does maybe a few hundred writes *per year*.
- So SQLite has headroom of several orders of magnitude for this use case.

And the senior doesn't just say "it'll be fine". They **write down the trigger** for changing their mind: "Move to Postgres if we ever need more than one app instance (high availability) or concurrent writes become a bottleneck." Then they make that move cheap by keeping data access behind repository interfaces (Phase D).

### Worked example: an ADR

```markdown
# ADR-001: Use SQLite as the primary database

Date: 2026-07-01
Status: Accepted

## Context
Single-user/family scale self-hosted password manager. Runs locally via
Docker Compose. One operator. Expected load: < 10 req/s, < 100k rows.
Durability matters a lot; availability matters little. The project is
also a portfolio piece, so setup must be one command.

## Decision
Use SQLite (pure-Go driver, no CGO) in WAL mode with foreign keys on,
behind repository interfaces.

## Alternatives considered
- **Postgres:** more scalable, richer types, but adds a service to run,
  configure, back up and secure. No current requirement needs it.
- **Embedded KV store (bbolt):** simple, but we want relational
  constraints (foreign keys, unique email) and SQL migrations.

## Consequences
- + Zero extra services; backup is `VACUUM INTO` a file.
- + Static binary, simple Docker image.
- − Single writer; no multi-instance deployment.
- − Must enable pragmas explicitly (foreign keys are OFF by default!).

## Revisit when
We need more than one app instance, or write latency under load exceeds
100 ms at p99.
```

Half a page, and in a year you'll know exactly why SQLite was chosen, what was considered, and when to reconsider. The "Revisit when" line is the most senior part: it turns a decision into a *conditional* decision.

> **Junior vs senior**
>
> A **junior** picks the stack they saw in a tutorial or the one that's trending: "Microservices with Kafka and Postgres on Kubernetes, because that's what real companies use."
>
> A **senior** picks the stack from constraints and writes down what would change their mind. They're comfortable choosing "boring" because they know complexity is a cost that's paid every day, while scale is a problem you might have someday.

### Exit criteria
- NFRs are written down (even as a short list).
- Every major technology choice has a one-line justification tied to a constraint.
- ADRs exist for the one-way-ish choices (DB, auth scheme, crypto scheme).

---

## Phase D: Design the shape (architecture thinking)

**Goal:** decide how the code is organized, so that each part has one job, changes stay local, and the important rules are hard to break.

This is the phase where people most often ask "how did they *know* to use layered architecture?" So this section goes slowly.

### The questions a senior asks

- **What are the different reasons this code will change?**
- **What must each part know, and, more importantly, what must it *not* know?**
- **Where are the boundaries where data is translated** (HTTP ↔ Go, Go ↔ SQL)?
- **What do I need to be able to test without the rest?**
- **Is this design proportionate**, or am I over-engineering for a small project?

### How a senior arrives at a layered architecture

A senior doesn't start with "I'll use layered architecture because that's the pattern". They start with a list of **reasons the code will change**, then draw lines between them. Architecture is the result of that analysis, not the input.

For this server, list the things that will change independently:

| Reason to change | Examples | Who cares |
|---|---|---|
| **HTTP concerns** | URL paths, status codes, JSON field names, headers, request parsing, auth middleware | The API contract / client |
| **Business rules** | "Unknown email gets default KDF params", "password change rotates the security stamp", "revision date must match" | The product / security model |
| **Storage concerns** | SQL queries, GORM tags, indexes, migrations, SQLite pragmas | The database |

These three things change for completely different reasons, at different times. If they're mixed in one function, every change risks breaking the others. For example, changing a JSON field name shouldn't require touching a SQL query, and switching SQLite to Postgres shouldn't require touching the "rotate security stamp" rule.

So you separate them, and you get three layers:

```
   HTTP request
        │
        ▼
 ┌──────────────┐  Knows: HTTP, JSON, status codes, Gin, which user is calling (from token).
 │   handlers   │  Must NOT know: SQL, GORM, how passwords are hashed.
 └──────┬───────┘
        │ calls with plain Go types (DTOs, IDs, context)
        ▼
 ┌──────────────┐  Knows: business rules, security rules, domain errors.
 │   services   │  Must NOT know: HTTP status codes, Gin, SQL, GORM.
 └──────┬───────┘
        │ calls through a small interface it defines
        ▼
 ┌──────────────┐  Knows: GORM, SQL, tables, how to translate DB errors to domain errors.
 │ repositories │  Must NOT know: HTTP, business rules.
 └──────┬───────┘
        ▼
     SQLite
```

The "must NOT know" lines are the actual design. A layer that doesn't know about HTTP *can't* accidentally return a 500 with a SQL error message in it.

### Why exactly these layers, and not more or fewer?

- **Fewer** (handlers talking directly to GORM): fine for a 200-line toy. But then business rules live in HTTP handlers, you can't test a rule without spinning up HTTP and a DB, and security rules get duplicated across endpoints. For a security product, centralizing the rules in one testable layer is worth the extra files.
- **More** (hexagonal / clean architecture with ports, adapters, use cases, entities, presenters…): these are valid patterns for large systems with many teams and many adapters. For one developer with one DB and one transport, the extra layers are ceremony that slows you down without protecting anything.

The senior's rule: **add a layer only when it separates two things that change for different reasons.** Three reasons to change, three layers.

> **When would a senior *not* use layers?** A 50-line CLI tool, a one-off script, a prototype that's going to be thrown away. Architecture is proportional to how long the code lives and how many reasons it has to change.

### Dependency direction and interfaces

The arrows point **down** only: handlers depend on services, services depend on repositories. Never the reverse.

But here's a subtlety that seniors care about: should the service depend on the *concrete* `UserRepository` struct? If it does, every service test needs a real database.

Idiomatic Go solves this with **small interfaces defined by the consumer**:

```go
// in package services: the service says what it NEEDS, nothing more
type UserStore interface {
    Create(ctx context.Context, u *models.User) error
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type AccountService struct {
    users UserStore // could be the real repo, or a fake in tests
}
```

*Why the consumer defines it:* the service only needs two methods, so its interface has two methods. The repository doesn't need to know the interface exists; Go interfaces are satisfied implicitly. This is the Go proverb "accept interfaces, return structs", and it's the single biggest enabler of easy testing.

*Why not define one big interface next to the repository?* Because then every consumer depends on every method, and every fake has to implement all of them. Small interfaces keep tests small.

### DTOs vs entities: the boundary is a security control

Juniors often see DTOs as boilerplate: "Why copy fields from `User` into `GetUserDTO` when they're the same?" In a password manager, the answer is a security one:

- The **entity** (`models.User`) is the database's view. It will hold `MasterPasswordHash`, `ProtectedUserKey`, `SecurityStamp`.
- The **response DTO** is the API's view. It must never contain those.

If handlers return entities, then one day someone adds a field to `User`, forgets `json:"-"`, and the hash ships in every API response. If handlers return DTOs built by **explicit field-by-field mapping**, adding a field to `User` changes nothing in the API. That's principle 1.8 again: **make the wrong thing hard**.

This is exactly why the `phase0/ci` review flagged `models.GetUserDTO(user)` (a struct conversion). It looks like a cleanup, but it couples the API shape to the DB shape, which removes the very protection DTOs exist to provide.

### Error design is architecture

Errors cross every layer, so they need a design just like data does. The senior approach:

```
 repository: gorm.ErrRecordNotFound  ──translate──►  apperr.ErrNotFound
 service:    returns apperr.ErrNotFound / ErrConflict / ErrUnauthorized (domain language)
 handler:    ONE function maps domain errors ──► HTTP status + stable code
                                                  logs the full cause server-side
                                                  sends the client {"code":"not_found"}
```

*Why:*
- Services speak the domain's language ("conflict"), not HTTP's ("409") or GORM's ("record not found").
- There's exactly one place that decides status codes, so they're consistent.
- Internal details (SQLite messages, table names) never reach the client (threat model: don't help attackers).
- The client gets a **stable** code it can program against, even if the message text changes.

### Package layout

```
server/
  cmd/server/main.go     wiring only: load config, build deps, start server
  internal/
    config/              typed config from env
    database/            open DB, pragmas, migrations
    models/              entities and DTOs
    repositories/        storage
    services/            business rules
    handlers/            HTTP
    middleware/          auth, request ID, logging, rate limiting
    apperr/              domain errors
  migrations/            versioned SQL files
```

*Why `internal/`:* the Go compiler forbids other modules from importing it. Your server's internals stay private, so you can refactor freely without breaking anyone.

*Why a thin `main.go`:* `main` can't be unit-tested. Keeping it to pure wiring (construct, connect, start) means everything with logic lives in testable packages.

*Why packages by layer and not by feature (`internal/vault/`, `internal/auth/`)?* Honestly, both are reasonable. By-feature scales better for big codebases with many features. By-layer is simpler to understand when starting out and fits a small service. It's a two-way door. A senior picks one, applies it consistently, and moves on.

### Worked example: one request through every layer

`PATCH /api/v1/ciphers/7f3c…` with an access token, updating an encrypted item:

```
1. middleware/auth      Validate JWT signature + expiry. Put userID in context.
                        (Knows: tokens. Doesn't know: ciphers.)
2. middleware/limits    Body ≤ 1 MiB, else 413.
3. handlers/cipher      Bind JSON into UpdateCipherRequest; validate shape
                        (binding tags). Read userID from context, cipherID
                        from URL. Call service. Map errors → status.
                        (Knows: HTTP. Doesn't know: SQL, revision rules.)
4. services/cipher      Rule: ciphertext format must be valid.
                        Rule: item must belong to userID → else ErrNotFound
                              (not "forbidden": don't reveal it exists).
                        Rule: request's revisionDate must equal stored one
                              → else ErrConflict (no silent overwrite).
                        Rule: bump user's revision date on success.
                        (Knows: rules. Doesn't know: HTTP, GORM.)
5. repositories/cipher  UPDATE ciphers SET … WHERE id = ? AND user_id = ?
                        Translate "0 rows" → ErrNotFound.
                        (Knows: SQL. Doesn't know: why.)
6. handlers/cipher      Map entity → CipherResponse DTO → 200.
```

Read the "Doesn't know" notes again. Each one is a bug that *can't happen* in that layer, because the layer doesn't have the information to make it.

> **Junior vs senior**
>
> A **junior** chooses layered architecture because the tutorial used it, then puts logic wherever it's convenient. "This check is easiest in the handler." Within months, rules are scattered across handlers and repositories, and nobody knows where "the" ownership check lives.
>
> A **senior** derives the layers from *reasons to change*, defines what each layer must *not* know, and treats violations as bugs in review. They can explain why the architecture has three layers and not five.

### Exit criteria
- You can name each layer's responsibility and what it must not know.
- Error flow and DTO boundaries are decided.
- Package layout is decided (and you've accepted it as a two-way door).

---

## Phase E: Design the contracts before the code

**Goal:** define the data model and the API precisely enough that implementing them is mostly mechanical.

A "contract" is anything another part of the system depends on: the database schema (migrations depend on it), the HTTP API (the client depends on it), the crypto formats (stored data depends on them). Contracts are one-way-ish doors (principle 1.2), so they get designed on paper first.

### The questions a senior asks

**Data model**
- What entities exist, and which threat (Phase B) shaped each one?
- What does the server need to *read*, and what can be opaque?
- What are the IDs, and are they safe to expose?
- What are the uniqueness, foreign key and deletion rules?
- How will the schema change over time?

**API**
- What are the resources, and who is allowed to touch each one?
- How does the caller identify *themselves* vs *other objects*?
- What does every error look like?
- How will the API evolve without breaking the client?

### Deriving the data model from the threat model

A senior doesn't start the schema by listing "things a password manager has". They start from the security model and ask what the server *needs to store* to do its job, and nothing more.

```
users
  id                    UUIDv7      ← T6: sequential ints leak user count and invite enumeration
  email                 unique, lowercased  ← "A@x.com" and "a@x.com" must be the same account
  email_verified        bool
  kdf_type, kdf_iterations, kdf_memory, kdf_parallelism
                                    ← the client needs these to derive keys (sent by prelogin)
  master_password_hash  Argon2id(auth hash) in PHC format
                                    ← T1: a stolen DB doesn't give a usable login credential
  protected_user_key    ciphertext  ← key wrapping: password change re-wraps only this
  security_stamp        random      ← rotated on password change → invalidates all sessions
  revision_date, created_at, updated_at

refresh_tokens
  id, user_id (FK, cascade)
  token_hash            SHA-256     ← T1: never store the token itself
  family_id                         ← reuse detection revokes the whole family
  expires_at, revoked_at, device_name, created_at

folders
  id, user_id (FK, cascade)
  name                  ciphertext  ← even folder names can be sensitive ("Bank", "Lawyer")
  revision_date

ciphers
  id, user_id (FK, cascade), folder_id (FK, set null)
  type                  enum: login | note | card | identity
  data                  ciphertext blob  ← the server can't read it, so don't give it columns
  favorite, deleted_at (soft delete), revision_date, created_at
```

Every field has an arrow pointing to the reason it exists. If you can't write the arrow, question whether the field should exist.

### The key insight: one encrypted blob instead of columns

A junior instinct is a `ciphers` table with `username`, `password`, `url` and `notes` columns, each encrypted. A senior asks: **"What does the server do with these columns?"** The answer is: nothing. It can't read them, search them, or validate them.

So separate columns give you only costs:
- Every new item type (card, identity, secure note, TOTP seed) needs a server migration.
- Column names leak structure ("this user has 40 items with a `card_number`").

One opaque `data` blob, whose structure only the client knows, means the client can evolve item types **without touching the server**. This is exactly why Bitwarden calls items "ciphers". The server keeps only what it genuinely needs for its own logic: owner, type (for limits), folder (for referential integrity), soft-delete state, and revision date (for sync).

> **The general lesson:** design the server's schema around what the server *does*, not around what the user *sees*.

### API design: let the URL prevent the bug

The most common serious API vulnerability is **Broken Object Level Authorization** (OWASP API Top 10, #1), also known as IDOR: `GET /users/42` works for user 41 too, because the server trusts the ID in the URL.

A senior doesn't just plan to "check carefully". They change the API so the bug has nowhere to live:

| Junior API | Senior API | Why |
|---|---|---|
| `GET /users/:id` | `GET /api/v1/accounts/me` | The "who" comes from the token. There's no ID to tamper with. |
| `GET /users` (list all) | *(doesn't exist)* | No user ever needs to list other users. |
| `GET /credentials?user_id=5` | `GET /api/v1/ciphers` | Scoped to the token's user, always. |
| `GET /ciphers/:id` | `GET /api/v1/ciphers/:id` + `WHERE user_id = <token user>` | Object IDs are unavoidable here, so ownership is enforced in the query, and a stranger's item returns **404**, not 403 (don't confirm it exists). |

Other contract decisions, each with its reason:

- **`/api/v1` prefix:** you can ship `/api/v2` later without breaking existing clients.
- **One error shape everywhere:** `{"code": "email_taken", "message": "…"}`. Clients switch on `code`; `message` is for humans and can change.
- **Status codes have fixed meanings:** 400 malformed, 401 not authenticated, 404 not found *or not yours*, 409 conflict, 413 too large, 422 well-formed but invalid, 429 rate limited. Write this table down once and never improvise.
- **Timestamps in UTC, RFC 3339.** IDs as strings (UUIDs).

### Auth flows as sequence diagrams

The auth flow is the riskiest contract in the system, so a senior draws it *completely* before coding it. Drawing reveals gaps, such as "what does prelogin return for an unknown email?"

**Registration**

```
CLI                                                         Server
 │ user types email + master password                         │
 │ choose KDF params (defaults)                               │
 │ masterKey = Argon2id(password, salt=email, params)         │
 │ authHash  = derive(masterKey, password)                    │
 │ userKey   = random 64 bytes                                │
 │ protected = encrypt(userKey, stretch(masterKey))           │
 │── POST /accounts/register {email, authHash, protected, kdf} ──►│
 │                          store Argon2id(authHash, random salt)  │
 │                          store protected, kdf params            │
 │◄──────────────────────────── 201 ──────────────────────────────│
```

**Login**

```
CLI                                                         Server
 │── POST /accounts/prelogin {email} ─────────────────────────►│
 │                         known email → its KDF params          │
 │                         unknown     → DEFAULT params (T6)     │
 │◄──────────────────── {kdf params} ───────────────────────────│
 │ masterKey = Argon2id(password, email, params)                │
 │ authHash  = derive(masterKey, password)                      │
 │── POST /auth/login {email, authHash} ───────────────────────►│
 │                         verify Argon2id, constant-time        │
 │                         wrong pw / unknown email → SAME 401   │
 │◄──── {accessToken (15 min), refreshToken, protectedUserKey} ─│
 │ userKey = decrypt(protectedUserKey, stretch(masterKey))      │
 │ (userKey lives in memory only)                               │
```

**Refresh with rotation and reuse detection**

```
CLI                                                         Server
 │── POST /auth/refresh {refreshToken R1} ────────────────────►│
 │                  look up hash(R1)                             │
 │                  valid & unused → revoke R1, issue R2         │
 │◄──────────────── {accessToken, R2} ──────────────────────────│
 │                                                               │
 │      (an attacker replays a stolen R1 later)                  │
 │                  hash(R1) found but already revoked            │
 │                  → REUSE DETECTED → revoke the whole family    │
 │                  → 401; the real user must log in again        │
```

Each arrow and each note becomes a test case later. That's not a coincidence; it's the point of drawing it.

### Worked example: one endpoint spec written before implementation

```markdown
## POST /api/v1/accounts/prelogin

Purpose: give the client the KDF params it needs to derive keys.
Auth: none (public). Rate limited: 10/min per IP, 5/min per email.

Request:
  { "email": "Alice@Example.com" }
  - email: required, valid format, ≤ 254 chars; normalized to lowercase

Response 200 (ALWAYS 200 for well-formed input):
  { "kdf": "argon2id", "iterations": 3, "memoryKiB": 65536, "parallelism": 4 }
  - known email → that user's params
  - unknown email → server default params (identical shape)

Errors:
  400 {"code":"invalid_request"}  malformed JSON / bad email format
  429 {"code":"rate_limited"}     too many requests

Security notes:
  - MUST NOT reveal whether the email exists (no 404, same shape, similar timing)
  - MUST NOT log the email at info level

Tests:
  - known email → its params
  - unknown email → defaults; response indistinguishable from a known
    email that uses defaults
  - "ALICE@example.com" and "alice@example.com" → same result
  - malformed email → 400
```

Notice the spec already contains the test list. When the time comes to write code (Phase G), the work is translating this into tests and then into code. Very little is left to figure out at the keyboard.

> **Junior vs senior**
>
> A **junior** designs the API by writing handlers and seeing what URLs come out. The API's shape is an accident of implementation, and its security properties (enumeration, IDOR) are whatever happened to fall out.
>
> A **senior** writes the contract first and designs security properties *into* it: `/me` instead of `/:id`, identical responses for unknown emails, 404 for other users' objects. They know that a contract, once a client depends on it, is very hard to change.

### Exit criteria
- `docs/design/api.md` lists every endpoint with auth, request, response and errors.
- The schema is sketched, with a reason for each field.
- Auth flows are drawn as sequence diagrams.
- You can explain the login flow on a whiteboard without notes.

---

## Phase F: When to start writing code, and how to start

**Goal:** know when design has done its job, then build a running skeleton before any features.

### When is "enough design" enough?

Design can become procrastination. Seniors stop designing when **the remaining unknowns are cheaper to answer in code than on paper.** In practice, start coding when:

1. You can explain the core flows (register, login, sync) without notes.
2. The one-way doors are decided and written down (security model, API shape, IDs, crypto formats).
3. The remaining questions are of the form "how does library X behave?" or "how slow is Argon2id with these params on my laptop?". Experiments answer those, not documents.
4. You can name the first five PRs.

If you're still debating whether the server should see the master password, it's too early. If you're debating whether to name the package `repo` or `repositories`, it's too late: you're bikeshedding. Start.

> **Proportionality:** for this project, Phases A-E are a few days of work at 1-2 h/day. A senior would not spend a month on documents for a solo project. The documents are short because their job is to make decisions, not to look thorough.

### Spikes: throwaway code to kill an unknown

A **spike** is a small, time-boxed experiment whose output is *knowledge*, not production code. For this project, sensible spikes:

| Unknown | Spike | What you learn |
|---|---|---|
| How slow is Argon2id? | A 20-line program timing `argon2.IDKey` with a few memory/iteration settings | Params that take ~0.5-1 s on your machine (the KDF default) |
| Does the encrypt/decrypt round-trip work? | Encrypt a string with AES-GCM (or AES-CBC+HMAC), decrypt it, flip a byte, confirm decryption fails | That you understand the API, nonces and authentication |
| Does the pure-Go SQLite driver enforce the pragmas? | Open a DB with `foreign_keys=ON`, insert an orphan row, confirm it fails | That foreign keys are actually enforced |

Rules for spikes:
- **Time-box it** (an hour or two). If it runs over, that's a finding in itself.
- **Throw it away.** Spike code was written to learn, not to last. Rewrite properly, with tests.
- **Write down what you learned** (a line in the design doc or an ADR).

### The walking skeleton

Before the first real feature, a senior builds a skeleton that does almost nothing but has **all the plumbing real**:

- A binary that starts from **config** (env vars, validated at startup).
- **Structured logging** with request IDs.
- A **DB connection** with the right pragmas, and a **migration** tool with one migration.
- **Graceful shutdown** with server timeouts.
- **One endpoint:** `GET /healthz`.
- **One test** that hits it via `httptest`.
- **CI** running lint, tests (with `-race`) and build on every PR.
- A **Makefile** (`make run`, `make test`, `make lint`).
- A **README** explaining how to run it.

*Why plumbing before features?* Because every feature built afterwards inherits the plumbing for free. If you add config, logging, error handling and CI *after* ten features, you retrofit all ten. If you add them first, feature one is already done properly.

It's also about **feedback loops**. With CI in place from PR #1, every later PR is checked automatically. Without it, you find out about broken tests whenever you happen to run them.

> **Compare with how this repo actually evolved:** it started with features (`/users` CRUD), and Phase 0 of ROADMAP retrofits the skeleton (layout, CI, config, logging, shutdown, error design, migrations). That's very common, and it's fine for a learning project. But notice the cost: features written before the skeleton have to be revisited (renaming `FindById`, changing `r.Run`, replacing `AutoMigrate`). A senior starting fresh would do it in the opposite order.

### Worked example: the first PRs a senior would make

```
PR 1  chore: init module, cmd/server/main.go, Makefile, .gitignore, README stub
      → `make run` serves GET /healthz

PR 2  ci: golangci-lint config + GitHub Actions (vet, lint, test -race, build)
      → every later PR is checked automatically

PR 3  feat: config from env (typed, validated) + slog JSON logging + request IDs
      → the binary behaves differently in dev/prod without code changes

PR 4  feat: http.Server with timeouts + graceful shutdown on SIGINT/SIGTERM
      → in-flight requests finish; Slowloris is mitigated

PR 5  feat: SQLite with pragmas + goose migrations + /readyz pings the DB
      → schema changes are versioned files, reviewed like code

PR 6  feat: apperr package + central error mapping + first handler test
      → every future endpoint gets consistent errors for free
```

Each PR is small (reviewable in 10-15 minutes), leaves `main` working, and makes the next PR easier. **There's no user-facing feature yet, and that's correct.** The first feature (registration) comes in PR 7, on top of a foundation that won't need revisiting.

> **Junior vs senior**
>
> A **junior** asks "when can I start coding?" on day one and starts with the most visible feature. Plumbing gets added "when we need it", which means it's retrofitted under pressure.
>
> A **senior** starts coding once the one-way doors are closed, runs spikes to kill technical unknowns, and builds a skeleton first. Their first week of code has no features, and their second week has features that never need rework.

### Exit criteria
- Spikes have answered the technical unknowns (and been deleted).
- The skeleton runs locally and in CI.
- The first feature PR can focus entirely on the feature.

---

## Phase G: How a senior knows what code to write

**Goal:** turn a designed feature into code methodically, so that "what do I write next?" always has an obvious answer.

This is the question juniors ask most: *"I know what the feature is supposed to do. How do I know which file to open and what to type?"* The answer is that seniors don't improvise this. They follow a repeatable process.

### The four questions for any piece of code

Before writing a function, a senior answers four questions, usually in their head, sometimes in a comment or a test:

1. **What goes in?** (Inputs, and which of them are untrusted.)
2. **What comes out?** (The result, and its type.)
3. **What can go wrong?** (Every failure: invalid input, not found, conflict, DB down, timeout.)
4. **What must never happen?** (The security and data invariants: "never store the plaintext", "never return another user's item", "never overwrite a newer revision".)

The answers to 3 and 4 are **the test list**. The answers to 1 and 2 are **the function signature**. Once you have both, the body is often the easiest part.

### Outside-in vs inside-out

There are two ways to build a vertical slice:

- **Inside-out:** start at the bottom (migration, repository), then service, then handler. Good when storage is the hard, uncertain part.
- **Outside-in:** start at the contract (the handler test from the API spec), then the service it needs, then the repository the service needs. Good when the *behavior* is the hard part, which, in a security product, it usually is.

A senior usually works **outside-in from the contract, but implements the core rules first**:

```
 1. Write down the test list from the endpoint spec (Phase E did most of this)
 2. Define the service's method signature and the interface it needs
 3. Write service unit tests with a fake store   ← the rules live here
 4. Implement the service until the tests pass
 5. Write the migration + repository (+ a repo test against real SQLite)
 6. Write the handler + a handler test via httptest
 7. Wire it in main.go / routes
 8. Run the whole thing manually once; then self-review the diff
```

Why the service first? Because that's where the *rules* are, and rules are what can be wrong in dangerous ways. HTTP binding and SQL are comparatively mechanical.

### Worked example: building "register" from a blank file

Here's the thinking for `POST /api/v1/accounts/register`, step by step.

**Step 1: The four questions**

1. *In:* email, auth hash, protected user key, KDF params. **All untrusted.**
2. *Out:* success (201) or an error.
3. *Can go wrong:* malformed JSON; invalid email; KDF params too weak (a malicious or buggy client sending `iterations=1`); auth hash of the wrong length; email already taken; DB error.
4. *Must never happen:*
   - storing the auth hash as received (it must be re-hashed with Argon2id);
   - two accounts for `A@x.com` and `a@x.com`;
   - the response containing any hash or key;
   - SQLite's error text reaching the client.

**Step 2: The test list** (straight from step 1)

```
service tests (fake store):
  ✓ valid input → user created; stored hash ≠ input hash; stored hash verifies against input
  ✓ email is lowercased before storing and before the uniqueness check
  ✓ existing email → ErrConflict (email_taken)
  ✓ KDF params below the minimum → ErrValidation
  ✓ store error → wrapped error, not swallowed

handler tests (httptest):
  ✓ valid request → 201, body has no hash/key fields
  ✓ malformed JSON → 400 invalid_request
  ✓ duplicate email → 409 email_taken, no SQLite text in the body

repository test (real in-memory SQLite):
  ✓ unique index on email rejects a duplicate (translated to ErrConflict)
```

Writing the list takes ten minutes. They're the most valuable ten minutes of the feature, because the list is the complete definition of "done".

**Step 3: Signatures and the interface**

```go
// services: what the service needs from storage (consumer-defined, small)
type UserStore interface {
    Create(ctx context.Context, u *models.User) error
}

// services: the method the handler will call
func (s *AccountService) Register(ctx context.Context, in RegisterInput) error
```

Notice what's *absent*: no `*gin.Context`, no `*gorm.DB`, no HTTP status. The service is pure business logic, which is why it's testable with a fake.

Also notice that `Register` doesn't check "does the email exist?" before inserting. Why? Because check-then-insert is a **race condition**: two simultaneous registrations can both pass the check. The unique index in the DB is the real guard, and the repository translates the constraint violation into `ErrConflict`. A senior asks *"what if two of these run at once?"* about every check-then-act.

**Step 4: Implement the service until the tests pass.** Normalize the email, validate KDF minimums, hash the auth hash with Argon2id and a random salt, build the entity, call `Create`.

**Step 5: Migration and repository.** Create the `users` table with a unique index on `email`. `Create` uses `db.WithContext(ctx)` and translates the unique-constraint error into `apperr.ErrConflict`.

**Step 6: Handler.** Bind JSON into a request DTO with `binding` tags, convert it to `RegisterInput`, call the service, map errors through the central error mapper, return `201`.

**Step 7: Wire it up.** Construct the repo, service and handler in `main.go`, and register the route.

**Step 8: Self-review before opening the PR.** Read your own diff as if someone else wrote it:
- Does any response include a field it shouldn't?
- Is any error ignored (`_ =`)?
- Is anything logged that shouldn't be (emails at info level, request bodies on auth routes)?
- Are there "remember to…" rules that the code could enforce instead?
- Is there dead or commented-out code?
- Would I understand this in six months?

The result is one small PR with one working, tested behavior. The next slice (prelogin) starts from the same process.

### Breaking a feature into PR-sized slices

"Authentication" is not one PR. A senior splits it so that each PR is **independently correct and mergeable**:

```
auth feature
 ├─ PR: users migration + repository (+ repo tests)
 ├─ PR: register endpoint (service + handler + tests)
 ├─ PR: prelogin endpoint (enumeration-safe)
 ├─ PR: login → access token only
 ├─ PR: refresh tokens (store the hash, issue on login)
 ├─ PR: refresh rotation + reuse detection
 ├─ PR: auth middleware + /accounts/me
 └─ PR: logout + change password (security stamp)
```

Rules for a good slice:
- **It leaves `main` working.** Never merge half a feature that breaks something.
- **It's reviewable in about 15 minutes.** Bigger PRs get worse reviews, not better ones.
- **It has one reason to exist.** "Add register endpoint": yes. "Add register endpoint, rename some files and upgrade Gin": no.

### Deciding what to test: weight tests by risk

Seniors don't aim for "100% coverage". They aim for **confidence where it matters**:

| Code | Risk if wrong | How much to test |
|---|---|---|
| Auth: login, token rotation, password change | Account takeover | Heavily: every branch, every failure, every "must never" |
| Authorization: every vault query | Cross-user data leak | One "user B can't touch user A's item" test **per endpoint** |
| Sync/conflict logic | Silent data loss | Heavily, including concurrent scenarios |
| Validation of untrusted input | Abuse, corruption | Table-driven tests + fuzzing for parsers |
| Error mapping | Info leaks, wrong codes | One table-driven test over all domain errors |
| Plain getters, wiring in `main` | Low | Covered indirectly; no dedicated tests |

The question is always: **"If this were wrong, how bad would it be, and would anything else catch it?"**

### "Designed for testing" vs "tested afterwards"

The fake store in step 3 only works because the service depends on an interface, which was decided in Phase D. That's the real reason for the architecture: **testability is a design property, not something you add later.** If a function is hard to test, that's design feedback. It usually means the function does too many things, or depends on concrete types it shouldn't.

> **Junior vs senior**
>
> A **junior** opens a file, writes code until it seems to work, tries it with curl, and then wonders how to test it. The tests (if any) check that the code does what it does, not what it should do.
>
> A **senior** writes down what can go wrong and what must never happen *first*. That list becomes the tests, the tests force a testable design, and the code is written last to satisfy both. "What do I write next?" is never a mystery: it's the next failing test.

### Exit criteria (per feature)
- Every item on the test list has a test.
- The PR is small, self-reviewed, and leaves `main` green.
- No "must never happen" rule depends on someone remembering it.

---

## Phase H: Build features in risk order

**Goal:** sequence the features so that each one builds on a trustworthy foundation, and the riskiest work gets the most time and attention.

### The questions a senior asks

- **What depends on what?** (You can't scope vault queries by user until you know who the user is.)
- **What's riskiest, and does it have to come early?**
- **What can be added later without changing anything earlier?** (Those features can safely wait.)
- **At each step, is the system still correct, even if incomplete?**

### The sequence, and why

```
 skeleton ─► accounts & auth ─► vault + authorization ─► sync & concurrency
                                                               │
           client ◄─ operations ◄─ hardening (rate limit, 2FA, email) ◄┘
```

| Order | Area | Why it comes here |
|---|---|---|
| 1 | **Accounts and auth** | Everything else needs "who is calling". It's also the riskiest code, so it gets done while you're fresh and have time to test it heavily. |
| 2 | **Vault + authorization** | The core value. Built on top of auth, so every query can be scoped by the token's user from day one, never retrofitted. |
| 3 | **Sync and concurrency** | Needs vault data to exist. It's the main source of *data loss* risk, so it comes before polish. |
| 4 | **Hardening** (rate limiting, 2FA, email, audit log) | These *add* defenses around correct code. They can be layered on (mostly as middleware and extra steps) without changing earlier contracts. Adding rate limiting to a broken login doesn't fix it. |
| 5 | **Operations** (metrics, backups, Docker, TLS) | Needs a working system to operate. Backups especially must be done before anyone stores real data. |
| 6 | **Full client** | The crypto core of the client exists from step 1 (it's needed to test auth). The polished CLI comes once the API is stable, so it isn't chasing a moving target. |

The general rule: **correctness before defenses, defenses before polish.** Each step leaves a system that's incomplete but never *wrong*.

### Key senior questions and traps, area by area

Concrete tasks for each area are in ROADMAP Phases 2-7. Here's the thinking that should accompany them.

**Accounts and auth**
- *"Can an attacker learn whether an email has an account?"* Check prelogin, login, and the timing of both. Trap: returning 404 for unknown emails, or skipping the Argon2 verify for unknown users (which makes them measurably faster).
- *"What happens to existing sessions when the password changes?"* Trap: changing the password but leaving old refresh tokens valid. The security stamp exists for this.
- *"If a refresh token leaks, how would I know?"* Trap: refresh without rotation, where a stolen token works forever.
- *"Am I comparing secrets in constant time?"* Trap: `==` on hashes.

**Vault and authorization**
- *"For this query, what stops user B from reading user A's row?"* There must be a concrete answer for **every** query: a `WHERE user_id = ?`. Trap: fetching by ID, then checking ownership in the handler, and forgetting one endpoint.
- *"What can I validate about data I can't read?"* Size, format, type, count. Trap: accepting a 500 MB "item".
- *"What happens on delete?"* Trap: hard deletes with no undo, or deleting a user and leaving their ciphers orphaned because SQLite foreign keys are off by default.

**Sync and concurrency**
- *"Two devices edit the same item offline. What happens?"* Trap: last-write-wins, which silently loses data. The fix is optimistic concurrency (revision dates, 409 on mismatch).
- *"What if two requests for the same user arrive at once?"* Trap: read-modify-write without a transaction, or check-then-act races.

**Hardening**
- *"What's the cheapest attack, and what does it cost the attacker after this change?"* Rate limiting turns millions of guesses per hour into dozens.
- *"Does this defense create a new attack?"* Trap: per-account lockout that lets anyone lock *you* out by spamming your email. Seniors combine per-IP limits, progressive delays and temporary (not permanent) lockout.
- *"What if the email provider is slow?"* Trap: sending email synchronously inside login, so SMTP latency becomes login latency.

**Operations**
- *"Have I ever restored a backup?"* Trap: backups that silently don't work.
- *"How would I know the system is broken before a user tells me?"* That's what metrics and health checks are for.

> **Junior vs senior**
>
> A **junior** builds features in the order they're excited about, or in the order the UI shows them, and adds "security stuff" at the end.
>
> A **senior** orders features by dependency and risk: auth first because everything depends on it and it's the most dangerous; defenses after correctness because defending broken code is pointless; polish last. At every point, the system is incomplete but never unsafe.

### Exit criteria
- Each area's "must never happen" list has tests.
- Each area's traps have been explicitly checked, not assumed.

---

## Phase I: Quality as a system, not an afterthought

**Goal:** make quality automatic, so it doesn't depend on discipline or memory.

### The questions a senior asks

- **What mistakes can a machine catch, so humans don't have to?**
- **What's the fastest feedback loop for each kind of mistake?**
- **What does "reviewed" actually mean on this project?**

### The feedback loop ladder

Seniors push each kind of mistake down to the cheapest place to catch it:

```
 cheapest, fastest
   ▲  compiler / types        "can't call a method that doesn't exist"
   │  formatter (gofmt)       "no style debates, ever"
   │  linter (golangci-lint)  "ignored error", "shadowed var", "weak crypto" (gosec)
   │  unit tests              "rule X holds"
   │  integration tests       "the layers fit together with a real DB"
   │  CI on every PR          "all of the above, on a clean machine, every time"
   │  code review             "is this the right design? is anything missing?"
   ▼  production / users      "it broke" (the most expensive place to find out)
 most expensive, slowest
```

Everything a machine can check (formatting, unchecked errors, known-insecure patterns) should be checked by a machine, so human review can focus on what machines can't judge: design, missing cases, naming, clarity.

### The tools, and the reason for each

- **`gofmt` / `goimports`:** formatting is decided once and never discussed again.
- **`golangci-lint`** with `errcheck` (ignored errors), `staticcheck` (bugs and simplifications), `gosec` (security smells like weak random numbers or SQL string building), `errorlint` (correct `errors.Is`/`As` use), `revive` (style, doc comments).
- **`go test -race`:** an HTTP server is concurrent (a goroutine per request). Data races are intermittent and nearly impossible to debug by hand, and the race detector finds them.
- **`govulncheck`:** tells you when a dependency you *actually call* has a known vulnerability.
- **Fuzzing (`go test -fuzz`):** for anything that parses untrusted input (ciphertext format, tokens). Fuzzers find the inputs you didn't think of.
- **CI with branch protection:** `main` only accepts PRs whose checks are green. This turns all of the above from "good intentions" into "physically impossible to skip".

### The test pyramid for this project

```
          ╱╲          few:  end-to-end with the Go client
         ╱  ╲               (register → login → create item → sync → logout)
        ╱────╲
       ╱      ╲       some: handler + real SQLite (httptest)
      ╱        ╲            cross-user access tests live here, one per endpoint
     ╱──────────╲
    ╱            ╲    many: service unit tests with fakes
   ╱              ╲         every rule, every failure, table-driven
  ╱────────────────╲
```

*Why this shape:* unit tests are fast and pinpoint failures; end-to-end tests are slow and vague when they fail but prove the whole thing works. You want lots of the first and a few of the second.

### The code review mindset

A senior reviewing a diff goes roughly in this order, from most to least important:

1. **Correctness:** does it do what it claims? What inputs break it?
2. **Security boundaries:** does untrusted input get validated? Is any query unscoped? Is anything sensitive logged or returned?
3. **Design:** is logic in the right layer? Does it follow the project's existing patterns? Does it couple things that should be separate?
4. **Tests:** do the tests check the *rules*, including failures, or only the happy path?
5. **Readability:** names, comments that explain *why*, no dead code.
6. **Style:** only what the linter didn't catch (which should be very little).

Concrete examples from this project's own `phase0/ci` review:

| Finding | Which level | The general lesson |
|---|---|---|
| `branch:` instead of `branches:` in the workflow | Correctness | Config is code. Verify it actually ran ("the check is green on GitHub"), don't assume. |
| `models.GetUserDTO(user)` struct conversion | Design / security boundary | A "simplification" that couples the API to the DB schema removes the protection DTOs exist for. |
| `log.Fatal("Error...")` | Correctness (operability) | An error message without the error is useless at 3 a.m. Always include the cause. |
| Commented-out code left in | Readability | Git remembers. Delete it. |
| `// FindAll find all` doc comments | Readability | A comment that repeats the name satisfies the linter but tells the reader nothing. Say *what you can't see from the signature*. |

### Commit and PR hygiene

- **One logical change per commit and per PR.** Easy to review, easy to revert.
- **The PR description explains *why*.** The diff shows *what*. Future you reads PR descriptions to understand decisions.
- **Conventional commits** (`feat:`, `fix:`, `refactor:`, `ci:`) make history scannable.
- **Review your own diff before asking anyone else.** You'll catch a third of the issues yourself.

> **Junior vs senior**
>
> A **junior** thinks of quality as "being careful": running tests when they remember, formatting by hand, hoping reviewers catch things.
>
> A **senior** thinks of quality as *a system*. They automate everything a machine can check, block merges on it, and spend human attention only on the things that need judgment.

### Exit criteria
- CI runs format, lint, race tests and vuln checks on every PR, and blocks merges.
- Tests follow the pyramid, with cross-user tests for every vault endpoint.

---

## Phase J: Operate it

**Goal:** be able to run the system, notice when it's unhealthy, and recover from the bad day.

### The questions a senior asks

- **How do I know it's working right now?**
- **How will I find out it's broken before a user does?**
- **What's the worst realistic bad day, and what do I do on it?**
- **Could someone other than me run this from the docs alone?**

### Design for the bad day

Seniors list bad days in advance and make sure each one has an answer:

| Bad day | Answer | Built in which phase |
|---|---|---|
| The disk dies | Backups via `VACUUM INTO` (safe while running), stored elsewhere, **and a tested restore** | J |
| A bad migration ships | Versioned migrations; back up before migrating; a documented rollback | F, J |
| The JWT signing key leaks | Key rotation with a `kid` header (two keys valid during the switch); the security stamp invalidates sessions | H, J |
| Someone brute-forces logins | Rate-limit metrics and logs show it; limits contain it | H |
| The server is slow or erroring | RED metrics (Rate, Errors, Duration) and structured logs with request IDs | F, J |
| A dependency has a CVE | `govulncheck` in CI flags it | I |

The most important line in that table is **"and a tested restore"**. A backup you've never restored is a hope, not a backup. Seniors schedule restore drills because the failure mode (discovering your backups are broken *during* the disaster) is so severe.

### Observability, briefly

- **Logs** answer "what happened in this specific request?" They're structured (key/value), carry request IDs, and *never* contain secrets, tokens or auth request bodies.
- **Metrics** answer "how is the system doing overall?" Request rate, error rate, latency, login failures.
- **Health checks** answer "should traffic be sent here?" `/healthz` means the process is alive; `/readyz` means it can serve (the DB is reachable).

### The runbook

A short `docs/runbook.md` covering: start, stop, upgrade (including migrations), back up, restore, rotate the signing key. Written so that someone who isn't you, or you in a year, can follow it under stress.

*Why write it down even for a solo project?* Because the bad day happens when you've forgotten the details. And for a portfolio, a runbook signals that you think about software as something that *runs*, not just something that compiles.

> **Junior vs senior**
>
> A **junior** considers the job done when the feature works on their machine.
>
> A **senior** considers it done when it can be run, observed, backed up, restored and upgraded by following written instructions, because that's what the software's life actually looks like after the first day.

### Exit criteria
- A restore from backup has been performed successfully at least once.
- Metrics and health checks exist; logs are structured and secret-free.
- The runbook exists, and you've followed it once from a clean machine.

---

## Phase K: Know when you're done

**Goal:** finish deliberately, with known trade-offs written down, rather than drifting into "almost done" forever.

### The questions a senior asks

- **Does the system keep the core promise from Phase A?** (Prove it: a test that dumps the DB and finds no plaintext.)
- **Is every "must never happen" covered by a test?**
- **What trade-offs did I accept, and are they written down?**
- **What would I do next, and why isn't it in v1?**

### Definition of done

ROADMAP Part 4 has the full production-readiness checklist. The senior addition is the **"known limitations"** section in the README. For example:

```markdown
## Known limitations (deliberate)
- Registration reveals whether an email is already registered (409).
  Accepted: the alternative (email-first signup) adds significant complexity.
- Single instance only (SQLite). See ADR-001 for when this changes.
- Forgotten master password = lost vault. This is inherent to zero-knowledge.
- The CLI must be trusted; a malicious client build could exfiltrate keys.
```

Writing limitations down is not admitting failure. It shows you *know* where the edges are, which is exactly what reviewers and interviewers look for. Every real system has limitations; only some engineers know theirs.

### What next

Keep a short "later" list (sharing, attachments, passkeys, import/export, password health reports) with one line on what each would require. That shows you've thought past v1 without letting it creep into v1.

> **Junior vs senior**
>
> A **junior** either never finishes ("just one more feature") or declares victory without checking the original goals.
>
> A **senior** goes back to the Phase A problem statement, verifies each promise with evidence, documents what was knowingly left out, and ships.

---

## Appendix A: Question bank

Print this. Before each phase, read its questions and write short answers.

### Understanding the problem
- [ ] Who uses this, and what problem do they have today?
- [ ] What's the one-sentence core promise?
- [ ] What would make users abandon it immediately?
- [ ] What's the minimum useful version?
- [ ] What's explicitly out of scope, now and forever?
- [ ] What are the consequences of the core promise (e.g. no password recovery)?

### Security
- [ ] What are the assets?
- [ ] Who are the attackers, and what can each one do?
- [ ] Where are the trust boundaries?
- [ ] If the DB is stolen, what can be read?
- [ ] If the server is compromised, what can be read?
- [ ] What residual risks am I accepting, and are they written down?

### Constraints and technology
- [ ] What scale, availability, durability and latency are actually needed? (Use numbers.)
- [ ] Who builds and maintains this, and how much time do they have?
- [ ] What does each technology choice cost, and which constraint justifies it?
- [ ] Under what conditions would I change this decision?

### Architecture
- [ ] What are the different reasons this code will change?
- [ ] What must each layer know, and what must it *not* know?
- [ ] What needs to be testable in isolation?
- [ ] Where are the translation boundaries (HTTP ↔ Go, Go ↔ SQL), and how do errors cross them?
- [ ] Is this proportionate to the project's size and lifespan?

### Contracts
- [ ] What does the server need to read, and what can be opaque?
- [ ] Are IDs safe to expose?
- [ ] For every endpoint, how does the server know who is calling and what they may touch?
- [ ] Does any endpoint reveal information to an unauthenticated caller?
- [ ] What does every error look like?
- [ ] How will the API and schema evolve?

### Before coding
- [ ] Are all one-way doors decided and written down?
- [ ] Are the remaining unknowns cheaper to answer in code?
- [ ] Which spikes do I need?
- [ ] Can I name the first five PRs?

### Every feature / function
- [ ] What goes in, and which inputs are untrusted?
- [ ] What comes out?
- [ ] What can go wrong?
- [ ] What must never happen?
- [ ] What if two of these run at the same time?
- [ ] What's the smallest PR that delivers this correctly?

### Every PR (self-review)
- [ ] Does any response or log contain something sensitive?
- [ ] Is every error handled or deliberately returned?
- [ ] Is every query scoped to the calling user?
- [ ] Do tests cover failures and "must nevers", not only the happy path?
- [ ] Is there dead code, commented-out code, or a "remember to…" rule the code could enforce?
- [ ] Would I understand this in six months?

### Operations
- [ ] How do I know it's healthy right now?
- [ ] Have I actually restored a backup?
- [ ] What's the plan if the signing key leaks?
- [ ] Could someone else run this from the docs?

---

## Appendix B: Templates

### B.1 ADR (Architecture Decision Record)

```markdown
# ADR-NNN: <decision in a short imperative phrase>

Date: YYYY-MM-DD
Status: Proposed | Accepted | Superseded by ADR-MMM

## Context
What situation forces a decision? What constraints apply? (Facts, not opinions.)

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

Keep ADRs in `docs/adr/`, numbered, and never edit an accepted one. Supersede it with a new one instead, so the history of *why* stays intact.

### B.2 Design doc

```markdown
# <Feature or system> design

## Goal
One paragraph. What problem, for whom.

## Non-goals
What this explicitly does not do.

## Background
What the reader needs to know first. Link to the threat model / ADRs.

## Design
How it works. Diagrams. Data model changes. API changes.

## Security considerations
Which threats does this touch? What must never happen?

## Alternatives considered
And why they lost.

## Testing plan
The test list, roughly.

## Open questions
What's still undecided.
```

Two to five pages. If it's longer, the feature is probably too big; split it.

### B.3 Endpoint spec

```markdown
## METHOD /api/v1/path

Purpose:
Auth: (none | access token) · Rate limit:

Request:   (fields, types, validation rules)
Response:  (status, body)
Errors:    (status + code, when)

Security notes:   (enumeration? ownership? logging?)
Tests:            (the list)
```

### B.4 PR description

```markdown
## What
One or two sentences.

## Why
The reason, with a link to the roadmap step / design doc / issue.

## How
Anything non-obvious about the approach. Alternatives you rejected.

## Testing
What tests were added; what you checked manually.

## Notes for the reviewer
Where you're unsure; what to look at first.
```

### B.5 Threat model row

```markdown
| # | Attacker | Capability | What they could get | Mitigation | Residual risk |
|---|----------|------------|---------------------|------------|---------------|
| Tn | who | what they can do | worst outcome | what stops it | what's left, and why accepted |
```

---

## Appendix C: Junior → senior anti-patterns

Each one with an example from this project's domain.

| # | Anti-pattern | Example here | Senior alternative | Why |
|---|---|---|---|---|
| 1 | **Starting from the tutorial shape** | `User{Password}` + `/users` CRUD on day one | Start from the core promise and the threat model | The tutorial shape is wrong for zero-knowledge and has to be thrown away |
| 2 | **Security as a later feature** | "We'll encrypt passwords in sprint 4" | Decide the security model before the schema | It's a one-way door; it changes the schema, API and auth |
| 3 | **Trusting IDs from the URL** | `PUT /users/:id` with no ownership check | `/me`; `WHERE user_id = <token user>` everywhere | IDOR is the #1 API vulnerability; design it out |
| 4 | **Returning entities** | `c.JSON(200, user)` | Explicit DTO mapping | One forgotten `json:"-"` leaks a hash |
| 5 | **"Simplifying" away a boundary** | `models.GetUserDTO(user)` struct conversion | Explicit field mapping in one helper | Couples the API shape to the DB shape |
| 6 | **Raw errors to clients** | `{"message": "UNIQUE constraint failed: users.email"}` | Domain errors → stable codes; log the details | Leaks internals; unstable contract |
| 7 | **Check-then-act** | `if exists(email) { return 409 }; insert()` | Unique index + translate the constraint error | Race condition: two requests both pass the check |
| 8 | **Silent overwrite** | Last write wins on cipher updates | Revision dates + 409 Conflict | Two devices silently destroy each other's edits |
| 9 | **Fast hashes for passwords** | `sha256(password)` | Argon2id with tuned params | GPUs try billions of SHA-256 guesses per second |
| 10 | **Storing tokens as-is** | `refresh_tokens.token = "abc…"` | Store `sha256(token)` | A DB leak would otherwise hand out live sessions |
| 11 | **Different errors for unknown users** | `404 user not found` on login | Identical 401 and similar timing | Account enumeration |
| 12 | **Plumbing last** | Adding config, logging and CI after ten features | Walking skeleton first | Retrofitting is more work than doing it once |
| 13 | **Huge PRs** | "Implement auth" in one 2,000-line PR | One slice per PR, ~15 min to review | Big PRs get rubber-stamped |
| 14 | **Testing only the happy path** | One test: "register works" | Test list from "what can go wrong / must never happen" | Bugs live in the failure paths |
| 15 | **Untestable design** | Service holds `*gorm.DB` directly | Consumer-defined interfaces + fakes | Testability is a design property |
| 16 | **Over-engineering** | Kubernetes, microservices, Kafka for a solo app | The simplest thing that meets the NFRs | Complexity is paid every day; scale might never come |
| 17 | **Under-engineering one-way doors** | `uint` auto-increment IDs "for now" | Decide ID type before the first migration | Changing IDs after data exists is painful |
| 18 | **Comments that repeat the code** | `// FindAll find all` | Explain what the signature doesn't show | Comments are for the reader, not the linter |
| 19 | **Keeping dead code "just in case"** | Commented-out blocks, unused `product.go` | Delete it; git remembers | Dead code misleads readers and may ship by accident |
| 20 | **Assuming instead of verifying** | "CI is set up" (but the workflow never ran) | Check the green tick; restore the backup; dump the DB | Untested assumptions fail at the worst moment |
| 21 | **Logging secrets** | Logging request bodies on `/auth/login` | Never log bodies on auth routes; structured fields only | Logs are copied, shipped and kept far longer than you think |
| 22 | **Unbounded input** | Accepting any body size | `MaxBytesReader`, item size and count limits | Cheap DoS and DB bloat |

---

## Appendix D: Mapping to ROADMAP and this repo

| Playbook phase | What it's about | ROADMAP section | Where this repo is (as of `phase0/ci`) |
|---|---|---|---|
| A: Understand the problem | Problem statement, scope | Decisions table, Context | Decisions made; no written problem statement yet |
| B: Must not go wrong | Threat model, security model | Part 2 (security model), Phase 1.1 | Zero-knowledge chosen; `security-model.md` is due in Phase 1 |
| C: Constraints and tech | NFRs, stack, ADRs | Decisions table | Stack chosen; ADRs not written yet (consider adding `docs/adr/`) |
| D: Architecture | Layers, interfaces, DTOs, errors | Part 1 review, Phase 0 (steps 1, 5, 6) | Layers and `internal/` exist; interfaces and `apperr` still to come |
| E: Contracts | Data model, API, auth flows | Phase 1 | Still to do (`api.md`, schema) |
| F: Start coding | Spikes, walking skeleton | Phase 0 | **In progress:** layout ✔, CI ✔ (pending fixes), config/logging/shutdown/migrations next |
| G: What code to write | Feature process, tests | Every phase's steps; Phase 0.9 | First tests are due in Phase 0 |
| H: Risk-ordered features | Auth → vault → sync → hardening → ops | Phases 2, 3, 4, 6 | Not started |
| I: Quality system | CI, linters, tests, review | Phase 0.8, Phase 5 | CI and linter in place |
| J: Operate | Observability, backups, runbook | Phase 6 | Not started |
| K: Done | Checklist, limitations | Part 4 | Not started |

### How to use the two documents together

- When starting a ROADMAP phase, first read the matching playbook phase and its questions in Appendix A. Write short answers before you write code.
- When you're stuck on "what do I write next?", go back to Phase G's four questions.
- When a review comment surprises you, find it in Appendix C. Most review comments are one of those 22 patterns.

---

*The goal isn't to memorize this document. It's to make its questions habitual, until you ask "what must never happen here?" without thinking about it. That habit is most of what "thinking like a senior" means.*
