# Password Manager Backend: Roadmap to Production

> Living document. Tick items off as you go. Detailed step-by-step guides for each phase live in `docs/guides/phase-N.md` and are written when you start that phase.

---

## Decisions (from your answers)

| Topic | Decision | What it changes |
|---|---|---|
| Security model | **Zero-knowledge** (client-side encryption) | The server stores only hashes and ciphertext. Crypto lives in the client. |
| Client | **Go CLI** (`client/` module) | Bearer tokens (no cookies). No CORS or CSP needed. The client's crypto package doubles as the API test harness. |
| Purpose | **Portfolio / learning project** | Keep SQLite. Invest in things reviewers notice: tests, CI, README, OpenAPI, design docs, clean PR history. |
| Sharing | **None** | No public/private keypairs, no organizations. Simpler schema. |
| Auth | **JWT access token + rotating opaque refresh token** | Phase 2 design. |
| Email | **Yes, via Mailpit locally** | A `Mailer` interface with an SMTP implementation. Mailpit runs in Docker as a fake inbox. A real provider later only needs config changes. |
| Deploy | **Local** (Docker Compose) | Phase 6 shrinks: no public TLS or reverse proxy. Backups and restore drills still matter. |
| Layout | **Standard Go layout** (`cmd/server` + `internal/`) | Done in Phase 0. |
| Working style | **You write, I review.** Each phase gets a detailed step-by-step guide (what to write + why), written when you start it. | Guides live in `docs/guides/phase-N.md`, based on your code at that moment so they never go stale. |
| Review flow | **One branch per step + GitHub PR** | I review the PR diff. CI (Phase 5) runs on the same PRs. |
| Pace | **1–2 h/day** | Rough estimates per phase below. Total is roughly 10–14 weeks. |
| Existing data | **Disposable** | Phase 0 resets the schema freely. |

---

## Context

You're building a Bitwarden-style password manager with Go, Gin, GORM and SQLite, starting with the backend. The codebase has a clean layered structure (routes → handlers → services → repositories), a working `/users` CRUD, and a `Credential` model with only a repository `Save`.

This roadmap takes you from there to a production-ready backend. For every step it explains **why**, because the goal is for you to learn to think like a senior engineer, not just to finish tasks.

**The most important idea in this document:** a password manager is a *security product first* and a CRUD app second. Almost every design choice below follows from one question: *"If the server, database or backups get stolen, what can the attacker read?"* For Bitwarden the answer is "nothing useful", and that property is called **zero-knowledge / end-to-end encryption**. It has to be decided *before* you build more features, because it changes the shape of your data model, your auth flow and your API.

---

## Part 1: Review of the current codebase

Here's what a senior reviewer would say about the code as it stands.

### What's good (keep doing this)
- **Layered architecture** (`handlers/` → `services/` → `repositories/`). Each layer has one job: HTTP concerns, business rules, data access. This will pay off when you add tests and auth.
- **DTOs separate from entities** (`CreateUserDTO`, `GetUserDTO`). The API contract isn't tied to the DB schema, so you can change one without breaking the other.
- **`json:"-"` on `Password`**. You're already thinking about what must never leave the server.
- **Mapping `gorm.ErrRecordNotFound` → 404**. That's the right instinct: translate storage errors into HTTP meaning.
- **Small, focused commits**. Good habit.

### Problems, ordered by severity

| # | Issue | Where | Why it matters |
|---|---|---|---|
| 1 | **No authentication at all.** Anyone can `GET /users`, read, update or delete any user by ID. | `routes/routes.go` | In a password manager this is the whole ballgame. Every non-public endpoint must know *who* is calling. |
| 2 | **`GET /users` lists every user.** | `routes.go`, `UserHandler.GetAllUsers` | Leaks your entire customer list (emails are PII). A normal user should only ever see *themselves* (`/me`). |
| 3 | **IDOR (Insecure Direct Object Reference).** `/users/:id` trusts the ID from the URL. | `userhandler.go` | Once auth exists, the user ID must come from the *token*, never from the URL. Otherwise user 5 can edit user 6. This is on the OWASP API Top 10 (#1: Broken Object Level Authorization). |
| 4 | **Passwords accepted but never stored, and no hashing plan.** | `services/userservice.go` | If you "fix" this by storing `dto.Password` directly, you store plaintext. A master password must never be stored, not even hashed with a fast hash like SHA-256. See Phase 2. |
| 5 | **`Credential.Password` is a plaintext column.** `GetCredentialDTO` returns it in plaintext. | `models/credential.go` | A DB leak would expose every user's every password. In the zero-knowledge design the server stores only *ciphertext* it cannot decrypt. |
| 6 | **Raw error messages go to clients** (`err.Error()`). A duplicate email returns SQLite's internal message. | `handlers/error.go` | Leaks internals (table names, driver details), which helps attackers. It also makes the API contract unstable. Map to stable error codes and log the details server-side. |
| 7 | **`CreateUser` returns the full `models.User` entity**, not a DTO. | `userhandler.go:27`, `userservice.go:10` | Inconsistent with the other endpoints, and the entity will grow sensitive fields (key material, hashes). Always return DTOs. |
| 8 | **Deleting a user doesn't delete their credentials.** SQLite doesn't enforce foreign keys unless `PRAGMA foreign_keys=ON`. | `database/connect.go` | Orphaned secret data survives account deletion, which is a privacy and legal (GDPR "right to erasure") problem. |
| 9 | **Sequential integer IDs** (`uint` primary keys). | `models/*` | They reveal how many users you have and make enumeration trivial. UUIDs are the norm for public IDs. |
| 10 | **Everything is hardcoded**: port `:8080`, DB path `test.db`, Gin in debug mode. | `main.go`, `connect.go` | You can't run different configs for dev, test and prod without editing code. See the 12-factor app "config in env" principle. |
| 11 | **`panic` on DB failure; no graceful shutdown.** | `connect.go`, `main.go` | When a server is stopped mid-request (every deploy), in-flight requests get cut off. Log and exit cleanly instead of panicking. |
| 12 | **`AutoMigrate` as the migration strategy.** | `connect.go` | Fine for prototyping. In production you need versioned, reviewable, reversible migrations, because AutoMigrate never drops or renames and can't tell you what it will do. |
| 13 | **Concrete types everywhere** (`*repositories.UserRepository` inside services). | `services/userservice.go` | You can't unit-test the service without a real DB. Small interfaces defined by the *consumer* are idiomatic Go and make testing easy. |
| 14 | **No `context.Context` propagation** (`db.WithContext(ctx)`). | repositories | If a client disconnects or a request times out, the DB query keeps running. Context is Go's standard for cancellation and deadlines. |
| 15 | **Small inconsistencies:** `UpdateUser` maps not-found to 400 (others use 404). `UpdateUserDTO` requires a `password` that's ignored. `ShouldBindJSON` and `ShouldBindBodyWithJSON` are mixed. Email comparison is case-sensitive (`A@x.com` ≠ `a@x.com`). | various | Each is small, but inconsistency is how bugs hide. Senior engineers pick one convention and enforce it. |
| 16 | **Housekeeping:** leftover `models/product.go`; `gin` marked `// indirect` in `go.mod`; no tests; no linter. | – | `go mod tidy`, delete dead code, add tooling (Phase 0). |

---

## Part 2: The security model (read this before Phase 1)

### How Bitwarden-style zero-knowledge works

```
            CLIENT (never sends the master password)                     SERVER
 master password ─┐
 email (salt) ────┴─► KDF (Argon2id, slow) ─► Master Key (256-bit)
                                                │
                     ┌──────────────────────────┼─────────────────────────┐
                     ▼                                                    ▼
      HKDF-expand → Stretched Master Key       PBKDF2(MasterKey, password, 1 iter)
                     │                                                    │
                     ▼                                                    ▼
   User Key (random 512-bit, generated once)          "Master Password Hash" ──► server re-hashes
   encrypted with Stretched Master Key                                              with Argon2id and
   = "Protected User Key" ─────────────────────────────────────────────────────►   stores ONLY that
                     │
                     ▼
   Every vault item encrypted client-side with User Key
   (AES-256-CBC + HMAC-SHA256, or AES-256-GCM)  ────────── ciphertext only ─────►  stores opaque blobs
```

**Why this design:**
- **The server never sees the master password or any plaintext secret.** A full database dump plus server compromise yields only ciphertext and slow-to-crack hashes. That is the product's core promise.
- **Why derive a separate "auth hash" from the master key?** The server has to verify you're you, but must not be able to decrypt your vault. The auth hash proves knowledge of the password without revealing the encryption key.
- **Why does the server hash the auth hash *again*?** If the DB leaks, the stored value isn't directly usable as a login credential (a "pass-the-hash" defense).
- **Why a random User Key wrapped by the Master Key, instead of encrypting items with the Master Key directly?** When the user changes their master password, you only re-encrypt *one small key*, not thousands of items. This is called **key wrapping / envelope encryption** and is used everywhere (AWS KMS, Signal, 1Password).
- **Why a slow KDF (Argon2id)?** Master passwords are human-chosen and weak. A slow, memory-hard KDF makes each brute-force guess expensive, even on GPUs.

**The consequence for the backend:** the server becomes a *mostly dumb, very careful storage and sync service*. It stores:
- the KDF params
- the server-side hash of the auth hash
- the protected User Key
- encrypted vault items
- revision timestamps

Its hard jobs are **authentication, authorization, rate limiting, sync correctness and operational safety**. Crypto on the *data* happens on the client.

**Consequence for "backend first":** you can't meaningfully test a zero-knowledge backend without *something* doing client-side crypto. The plan below adds a small Go test client (`client/` already exists!) early, in Phase 2. That gives you a way to exercise the API realistically, and it becomes the seed of your real client later.

> The alternative (server-side encryption with a server-held key) is simpler, but it is **not** Bitwarden-like: whoever controls the server can read everything. You chose zero-knowledge.

---

## Part 3: Phased plan

Every phase ends with a working, tested, committed state. **Don't start a phase until the previous one's "Definition of Done" holds.** That discipline is the difference between a project that ships and one that rots.

**Workflow for every step:** create a branch (`phase0/config`, `phase2/login`, …), keep the PR small (one step, reviewable in about 15 minutes), open a GitHub PR and ask me to review it. Fix the review comments, merge, repeat.
*Why:* small PRs get better reviews and are easy to revert. This is how real teams ship.

### Phase 0: Foundations and hygiene (~2 weeks)
*Goal: make the codebase safe to grow. No new features.*

1. **Housekeeping and layout.**
   - Run `go mod tidy` and delete `models/product.go`.
   - Move to the standard layout: `cmd/server/main.go` (wiring only) plus `internal/{config,database,handlers,services,repositories,models,…}`.
   - Add a `Makefile` with `run`, `test`, `lint` and `migrate` targets. Windows note: use `make` via Git Bash/Scoop, or a `Taskfile`.

   *Why:* dead code and wrong dependency metadata confuse future you. `internal/` is enforced by the Go compiler (other modules can't import it), and `cmd/` separates *what runs* from *what's reusable*. One-command workflows reduce friction.
2. **Configuration from environment.** Create a `config` package that loads `PORT`, `DB_PATH`, `ENV` (dev/prod), `LOG_LEVEL` and similar settings into a typed struct, and validates them at startup.
   *Why:* the same binary must run in dev, CI and prod. Failing fast on bad config beats failing at 3 a.m.
3. **Structured logging with `log/slog`** (standard library), plus a request-ID middleware.
   *Why:* in production you search logs as data (`user_id=… request_id=…`). Request IDs let you trace one request across log lines. **Rule: never log secrets, tokens or request bodies on auth routes.**
4. **Graceful shutdown.** Use `http.Server` + `signal.NotifyContext` + `srv.Shutdown(ctx)` instead of `r.Run()`. Set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout`.
   *Why:* deploys stop the process, and in-flight requests should finish. Timeouts protect against Slowloris-style attacks, and Go's default server has *none*.
5. **Error handling design.** Define domain errors in a package like `internal/apperr`: `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrValidation`. Services return these. One place (`handleError`) maps them to HTTP status plus a stable error `code`, and logs the underlying cause.
   *Why:* services shouldn't know about HTTP, and handlers shouldn't know about GORM. Clients get stable codes, and you keep the details.
6. **Interfaces for testability.** In `services`, define `type UserStore interface { … }` with only the methods the service uses. Pass `ctx context.Context` as the first parameter everywhere, and use `db.WithContext(ctx)`.
   *Why:* "accept interfaces, return structs" is idiomatic Go. It lets you unit-test services with in-memory fakes. Context enables cancellation and timeouts.
7. **Database setup done right.**
   - SQLite pragmas: `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000`, `synchronous=NORMAL`.
   - Switch from `AutoMigrate` to **versioned migrations** (`pressly/goose` or `golang-migrate`) with SQL files checked into git.
   - Primary keys become **UUIDs** (v7 is time-ordered, which suits indexes). Index `email` as unique on its lowercased form.

   *Why:* WAL allows concurrent reads during writes. Foreign keys make cascade delete work. Migrations make schema changes reviewable and repeatable across environments.
8. **Tooling and minimal CI.** Add `golangci-lint` (with `gosec`, `errcheck`, `staticcheck` enabled), `go test -race`, and `govulncheck`. Add a first GitHub Actions workflow that runs lint and tests on every PR.
   *Why:* machines catch whole classes of bugs for free. Senior engineers automate taste. CI starts now because you're working in PRs from day one (Phase 5 extends it).
9. **First tests.** Write service unit tests with fakes, plus one handler test using `httptest` + in-memory SQLite (`file::memory:?cache=shared`). Use table-driven tests.
   *Why:* you're about to rewrite the user model. Tests are what let you refactor without fear.

**Definition of Done:**
- `make lint test` is green.
- The server starts from env config and shuts down cleanly on Ctrl+C.
- A duplicate email returns `409 {"code":"email_taken"}` with no SQLite text.

### Phase 1: Design before code (~3–4 days)
*Goal: write down decisions. Seniors write short design docs; juniors jump to code.*

1. Write `docs/design/security-model.md`: threat model (who are the attackers: DB thief, malicious server operator, network attacker, stolen device?), what's encrypted where, KDF parameters, algorithms.
2. Write `docs/design/api.md`: resources, endpoints, auth scheme, error format, versioning (`/api/v1`).
3. Decide the data model (sketch below).

```
users            id(uuid) email(unique, lowercased) email_verified
                 kdf_type kdf_iterations kdf_memory kdf_parallelism
                 master_password_hash (server-side Argon2id of client auth hash)
                 protected_user_key (ciphertext string)
                 security_stamp (rotated on password change → invalidates sessions)
                 created_at updated_at revision_date

refresh_tokens   id user_id token_hash expires_at revoked_at device_name created_at
                 (store a HASH of the token, never the token itself)

folders          id user_id name(encrypted) revision_date
ciphers          id user_id folder_id type(login/note/card/identity)
                 data(encrypted blob) favorite deleted_at(soft delete / trash)
                 revision_date created_at
```

*Why a generic `ciphers.data` blob instead of `username` and `password` columns?* The server can't read the fields anyway, and one encrypted JSON blob lets the client add item types (cards, notes, TOTP seeds) **without server migrations**. This is exactly why Bitwarden calls them "ciphers".

**Definition of Done:** both docs exist, and you can explain the login flow on a whiteboard without looking.

### Phase 2: Accounts and authentication (~3 weeks)
*Goal: a user can register, log in and prove who they are on every request.*

1. **Reference crypto in `client/`.** Write a Go package that does KDF, key derivation and encrypt/decrypt, using `golang.org/x/crypto/argon2`, `crypto/hkdf`, `crypto/aes` and `crypto/hmac`. Write tests with known vectors.
   *Why:* you need a realistic caller for the API, and you'll reuse this for the CLI client later. **Never invent crypto primitives.** Only *compose* standard ones exactly as specified.
2. **`POST /api/v1/accounts/prelogin`** `{email}` → returns the KDF params. For unknown emails, return *default* params rather than 404.
   *Why:* the client needs the params to derive keys before logging in. Returning defaults for unknown emails prevents **account enumeration** (attackers learning who has an account).
3. **`POST /api/v1/accounts/register`** `{email, masterPasswordHash, protectedUserKey, kdf…}`. The server re-hashes `masterPasswordHash` with **Argon2id** (random per-user salt, stored in PHC string format).
   *Why Argon2id and not bcrypt?* It's the OWASP-recommended default, memory-hard, and has no 72-byte input limit.
4. **`POST /api/v1/auth/login`** → verify with a constant-time compare. Issue a **short-lived access token** (JWT, about 15 min, signed with EdDSA or HS256, containing `sub`, `exp`, `iat`, `jti` and `security_stamp`) plus a **long-lived opaque refresh token** (random 32 bytes, only its SHA-256 stored).
   *Why two tokens?* Access tokens are checked cheaply on every request without a DB hit. Refresh tokens are revocable because they're in the DB. Short access lifetime limits the damage from a leak.
5. **`POST /api/v1/auth/refresh`** with **rotation and reuse detection**: each refresh invalidates the old token. If a revoked token is ever reused, revoke the whole token family.
   *Why:* a reused refresh token means it was stolen. Rotation turns that theft into a detectable event.
6. **`POST /api/v1/auth/logout`** revokes the refresh token.
7. **Auth middleware**: validates the JWT and puts `userID` into `gin.Context`. **Remove** `GET /users` and `/users/:id`, and replace them with `GET/PATCH/DELETE /api/v1/accounts/me`.
   *Why:* it fixes issues #1–3 structurally. There's no ID in the URL to tamper with.
8. **Change master password** endpoint: requires the current hash, accepts the new hash plus the *re-wrapped* User Key, and rotates `security_stamp`, which kills all existing sessions.
   *Why:* this is the payoff of key wrapping. The vault items don't change at all.
9. **Delete account** requires re-entering the master password hash and cascades all data.
   *Why:* destructive actions need re-authentication, even with a valid token.

**Definition of Done:** an integration test (Go test client) can register → prelogin → login → call `/me` → refresh → logout, and the old refresh token then fails. Also:
- the DB contains no plaintext password
- a wrong password and an unknown email give identical responses and similar timing

### Phase 3: The vault (ciphers, folders, sync) (~2 weeks)
*Goal: store and sync encrypted items safely.*

1. **Folders and ciphers CRUD** under `/api/v1/folders` and `/api/v1/ciphers`. **Every repository query includes `WHERE user_id = ?`**, with the user ID taken from the token.
   *Why:* authorization must be enforced at the data layer, not just by "checking afterwards". Return 404 (not 403) for other users' items so their existence isn't leaked.
2. **Validate what you can't read.** The server can't validate plaintext, but it *can* enforce:
   - max blob size
   - the ciphertext format (e.g. `2.<iv>|<ct>|<mac>`)
   - allowed `type` values
   - item count per user

   *Why:* this is abuse and DoS protection, and it keeps garbage out of your DB.
3. **Soft delete (trash)** with `deleted_at`, plus restore and purge-after-30-days.
   *Why:* users delete things by accident. For a password manager, losing a credential can mean losing an account.
4. **Sync:** `GET /api/v1/sync` returns the profile, folders and ciphers. Every write bumps `users.revision_date`. The client asks "anything newer than X?"
   *Why:* multiple devices must converge. Revision dates are the simplest correct scheme.
5. **Optimistic concurrency:** updates send the `revision_date` they're based on. If it's stale, the server returns `409 Conflict`.
   *Why:* without this, two devices editing the same item silently overwrite each other ("last write wins" = data loss).
6. **Pagination and limits** on list endpoints.

**Definition of Done:** the test client can create, edit, trash, restore and sync items from two simulated devices, and a conflicting edit gets a 409. User A can't read, update or delete user B's items (explicit tests for each endpoint).

### Phase 4: Security hardening (~2–3 weeks)
*Goal: assume attackers are actively probing you.*

1. **Rate limiting** per IP and per account on `prelogin`, `login`, `register` and `refresh` (token bucket, e.g. `golang.org/x/time/rate` with an LRU of limiters). Add progressive delays and temporary lockout after N failures.
   *Why:* online brute force is the most common attack on login endpoints.
2. **Two-factor auth (TOTP)**: enrollment with QR secret plus confirmation code, recovery codes (hashed), and a login step that requires the code.
   *Why:* it defends against master password reuse and phishing.
3. **Email: verification and new-device notifications.** Define a `Mailer` interface (`Send(ctx, msg) error`) with an SMTP implementation, and run **Mailpit** in Docker as a local fake inbox (web UI at `localhost:8025`). Tests use an in-memory fake mailer. Send email *asynchronously* (a goroutine plus a small queue) so a slow SMTP server never slows down login.
   *Why:* the interface is the lesson. Your code depends on "something that sends mail", not on a vendor. Switching to Resend or SES later is a config change, which also impresses reviewers.
4. **HTTP hardening:**
   - security headers (`X-Content-Type-Options: nosniff`, `Cache-Control: no-store` on vault responses)
   - no CORS: the only client is the Go CLI, so leave CORS **disabled**. That's safer than a permissive config.
   - request body size limits (`http.MaxBytesReader`)
   - `gin.SetTrustedProxies` set correctly, so rate limiting sees real client IPs
5. **Audit/security event log**: logins, failed logins, password changes, 2FA changes, account deletion, each with IP, user agent and timestamp. The user can view their own events.
   *Why:* users and you need to detect account takeover.
6. **Secrets management**: the JWT signing key comes from env or a secret file, is never committed, and has a **rotation** plan (`kid` header supporting two active keys).
7. **Security testing:** `gosec`, `govulncheck` in CI, fuzz tests (`go test -fuzz`) on parsers and validators, and a manual pass against the **OWASP ASVS Level 2** checklist.

**Definition of Done:**
- The 100th rapid login attempt gets `429`.
- 2FA works end to end.
- The ASVS checklist is reviewed, with gaps written down.

### Phase 5: Quality, CI and portfolio polish (~1 week)
*Goal: every change is automatically checked, and the repo presents well.*

1. **Extend GitHub Actions** (started in Phase 0): add `-cover`, `govulncheck`, build, and a coverage badge. Turn on branch protection for `main` (CI must pass before merge).
2. **Test pyramid:** many service unit tests, fewer handler/integration tests against real SQLite, and a few end-to-end tests with the Go client. Track coverage on `services/` and the auth code specifically (aim for 80% or more there; coverage elsewhere matters less).
3. **OpenAPI spec** (hand-written or generated with `swaggo/swag`) checked into the repo.
   *Why:* it's the contract your future client is built against, and it doubles as documentation.
4. **Conventional commits and PRs, even solo.**
   *Why:* it trains the habit, and PR descriptions become your project history.
5. **README for reviewers:** what it is, the security model diagram, how to run it (`docker compose up`), API docs link, and design decisions and trade-offs.
   *Why:* for a portfolio project, the README is the first (and often only) thing a hiring manager reads. Explaining trade-offs is what signals seniority.

### Phase 6: Operations and local deployment (~1 week)
*Goal: run it like a real service, locally, and survive bad days.*

1. **Health endpoints:** `/healthz` (process alive) and `/readyz` (DB reachable).
2. **Metrics:** Prometheus `/metrics` with request count, latency, error rate and login failures.
   *Why:* you can't fix what you can't see. These are the "RED" metrics (Rate, Errors, Duration).
3. **Container:** a multi-stage Dockerfile that produces a static binary on a `distroless` or `scratch` image, running as a non-root user.
   *Why:* a small attack surface and reproducible builds. Pure-Go SQLite makes a static binary easy.
4. **`docker-compose.yml`**: the server, Mailpit, and optionally Prometheus + Grafana. One command brings up the whole stack.
   *Why:* anyone (including a reviewer) can run your project in a minute. That's reproducibility.
5. **TLS, even locally:** the server can serve HTTPS with a local cert (`mkcert`) and the CLI refuses plain HTTP unless given a `--insecure` flag.
   *Why:* even with end-to-end encryption, the tokens travel over the wire. Building "HTTPS by default" in now means you never ship without it.
6. **Backups:** a `make backup` target using SQLite's online backup (`VACUUM INTO 'backup-<date>.db'`), plus a documented **restore drill**. Litestream is optional.
   *Why:* a backup you haven't restored isn't a backup. Copying the `.db` file while the server writes to it can produce a corrupt copy, which is why you use `VACUUM INTO`.
7. **Stay on SQLite.** With WAL on one machine, SQLite handles far more than a personal project needs. Document *when* you'd move to Postgres (multiple app instances, HA), and note that the repository interfaces from Phase 0 contain that change. Writing down the trade-off is the senior part.
8. **Runbook** (`docs/runbook.md`): how to start, upgrade (run migrations), restore a backup and rotate the JWT signing key.

**Definition of Done:**
- `docker compose up` runs the whole stack over HTTPS.
- A backup restored into a fresh environment successfully.
- Metrics are visible.

### Phase 7: The Go CLI client (after the backend is done)
Your `client/` crypto package from Phase 2 grows into the real client. It would use `cobra` for commands, store tokens and the encrypted local cache in an OS keychain or file (`0600` permissions), and never write the decrypted User Key to disk. Commands would look like `login`, `sync`, `list`, `get`, `add`, `edit`, `generate` (password generator) and `lock`. We'll plan this in detail when you get there.

### Later / optional ideas
TOTP code storage in items, encrypted attachments, passkeys/WebAuthn as a second factor, import from Bitwarden/Chrome CSV, a password health report (reused/weak, computed *client-side*).

---

## Part 4: Production-readiness checklist (the finish line)

- [ ] No plaintext secrets anywhere: DB, logs, errors, backups
- [ ] Every endpoint is authenticated, except register/prelogin/login/refresh/health
- [ ] Every data query is scoped by the token's user ID; cross-user access has tests
- [ ] Rate limiting and 2FA on auth; generic errors (no enumeration)
- [ ] Versioned migrations; foreign keys on; cascade deletes verified
- [ ] Config from env; secrets not in git; key rotation documented
- [ ] Structured logs, metrics, health checks
- [ ] CI: lint, race tests, vuln scan; ≥80% coverage on auth and vault services
- [ ] HTTPS by default, CORS disabled, body size limits, server timeouts
- [ ] Backups plus a *tested* restore
- [ ] OpenAPI spec and runbook
- [ ] External review of the security design (even a knowledgeable friend is better than nobody)

---

## Part 5: Questions (answered)

All 10 questions are answered, and the answers are captured in the **Decisions** table at the top. I'll add new questions at the start of each phase guide, where the code at that point raises them.

