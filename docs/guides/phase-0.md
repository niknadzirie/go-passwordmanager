# Phase 0 Guide: Foundations and hygiene

> **Goal:** make the codebase safe to grow. No new features. At the end, the server has the skeleton that every production Go service has. Everything after this phase (auth, vault, rate limiting) plugs into it.
>
> **Time:** about 2 weeks at 1–2 h/day. **Steps:** 8, each one a branch and a PR.
>
> **How to use this guide:** read a whole step before starting it. Write the code yourself. Snippets show *shape* (signatures, structure) and leave the implementation to you, marked `// TODO`. When a step's "Check it works" passes, open a PR and ask me to review it.

## Contents

| Step | Branch | What you build | ~Time |
|---|---|---|---|
| 1 | `phase0/layout` | Standard layout, cleanup, Makefile | 1–2 days |
| 2 | `phase0/ci` | Linter + GitHub Actions | 1 day |
| 3 | `phase0/config` | Config from environment | 1 day |
| 4 | `phase0/logging` | `slog` + request ID + access log | 1–2 days |
| 5 | `phase0/server` | `http.Server`, timeouts, graceful shutdown | 1 day |
| 6 | `phase0/errors` | Domain errors + one error mapper | 1–2 days |
| 7 | `phase0/interfaces-tests` | Interfaces, `context`, first unit tests | 2 days |
| 8 | `phase0/database` | Pragmas, migrations, UUIDs, email normalization, integration test | 2–3 days |

**Why this order?** Each step makes the next one safer:
- CI comes second, so every later PR is checked automatically.
- Tests (step 7) come *before* the big database rewrite (step 8), so the rewrite is covered.

---

## Step 1: Standard layout and cleanup
**Branch:** `phase0/layout`

### What to do
1. Delete `server/models/product.go`.
2. Run `go mod tidy` in `server/`. `gin` moves from `// indirect` to a direct requirement.
3. Restructure `server/` like this:
   ```
   server/
   ├── cmd/
   │   └── server/
   │       └── main.go          ← only wiring: config → db → repos → services → handlers → router → run
   ├── internal/
   │   ├── database/
   │   ├── handlers/
   │   ├── models/
   │   ├── repositories/
   │   ├── routes/
   │   └── services/
   ├── go.mod
   └── Makefile
   ```
   Update the import paths: `passwordmanager-server/handlers` becomes `passwordmanager-server/internal/handlers`, and so on.
4. **Remove the credential code for now:** `models/credential.go` and `repositories/credential_repo.go`. Also remove `Credentials []Credential` from `User` and `&models.Credential{}` from `AutoMigrate`.
5. Remove the `Password` fields from `User`, `CreateUserDTO` and `UpdateUserDTO`. They're unused, and Phase 2 brings back a *correct* version (`master_password_hash`, stored as an Argon2id hash).
6. Add a `Makefile`:
   ```makefile
   .PHONY: run build test lint tidy

   run:
   	go run ./cmd/server

   build:
   	go build -o bin/server ./cmd/server

   test:
   	go test ./...

   lint:
   	golangci-lint run

   tidy:
   	go mod tidy
   ```
   Recipe lines must start with a **tab**, not spaces.

### Why
- **`internal/`** is enforced by the compiler: no other module (including your future `client/`) can import these packages. That keeps your server's internals private, so you can refactor them freely.
- **`cmd/server/main.go`** holds only *wiring*. Keeping `main` thin means everything important lives in packages you can test. `main` itself is untestable.
- **Why delete code you just wrote?** `Credential` stores a plaintext `Password` column, and the zero-knowledge design replaces it with encrypted `ciphers` (Phase 3). The `Password` fields are dead (never saved) and misleading. Keeping wrong code around "for later" is how it accidentally ships. Git remembers it if you ever want it back. Senior engineers delete code without sentiment.
- **Makefile**: one command per task means you, CI and anyone reading the repo run things identically.

### Common mistakes
- Forgetting to update an import. `go build ./...` finds these instantly.
- `make` isn't installed by default on Windows. Install it with `scoop install make` or `winget install ezwinports.make`, or run it from Git Bash. Alternatively, use [Task](https://taskfile.dev) with a `Taskfile.yml`. Either is fine; just pick one.

### Check it works
- `go build ./...` and `go vet ./...` pass.
- `make run` starts the server, and `curl localhost:8080/users` responds.
- Delete `test.db` first, because the schema changed. Your data is disposable.

---

## Step 2: Linter and CI
**Branch:** `phase0/ci`

### What to do
1. Install [golangci-lint](https://golangci-lint.run/welcome/install/) v2.
2. Add `server/.golangci.yml`:
   ```yaml
   version: "2"
   linters:
     default: standard        # errcheck, govet, ineffassign, staticcheck, unused
     enable:
       - gosec                # security issues
       - bodyclose            # unclosed HTTP bodies
       - errorlint            # correct use of errors.Is / errors.As
       - revive               # style
   formatters:
     enable:
       - gofmt
       - goimports
   ```
3. Add `.github/workflows/server.yml` at the **repo root**:
   ```yaml
   name: server
   on:
     pull_request:
     push:
       branches: [main]
   jobs:
     test:
       runs-on: ubuntu-latest
       defaults:
         run:
           working-directory: server
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version-file: server/go.mod
         - run: go vet ./...
         - run: go test -race ./...
         - uses: golangci/golangci-lint-action@v8
           with:
             version: latest
             working-directory: server
   ```
4. Run `make lint` locally and fix what it reports.
   You'll likely see `errcheck` on `r.Run(":8080")`. Don't silence it: step 5 replaces that line.

### Why
- **Linters catch real bugs**: ignored errors, unchecked type assertions, shadowed variables, security smells (`gosec`). They cost nothing once they're in CI.
- **CI on every PR** means "it works on my machine" stops being a thing. It's also the first thing a reviewer of your portfolio looks for.
- **`-race`** turns on Go's race detector. HTTP servers are concurrent (one goroutine per request), so data races are the nastiest bugs you'll hit, and this finds them.

### Common mistakes
- On Windows, `go test -race` needs a C compiler (cgo), so it may fail locally. Run `go test ./...` locally and let CI (Linux) run `-race`.
- Adding `//nolint` to make warnings go away. Only do that with a comment explaining *why* the warning is wrong in that spot.

### Check it works
Open the PR. The GitHub Actions check turns green.

---

## Step 3: Configuration from environment
**Branch:** `phase0/config`

### What to do
Create `internal/config/config.go`:

```go
package config

type Config struct {
	Env      string // "dev" or "prod"
	Port     string
	DBPath   string
	LogLevel slog.Level
}

// Load reads configuration from environment variables, applies dev defaults,
// and returns an error if anything is invalid.
func Load() (Config, error) {
	// TODO:
	// - read APP_ENV (default "dev"), PORT (default "8080"),
	//   DB_PATH (default "passwordmanager.db"), LOG_LEVEL (default "info")
	// - validate: Env must be "dev" or "prod"; Port must be a number 1–65535;
	//   LOG_LEVEL must parse (hint: slog.Level has UnmarshalText)
	// - return a single error listing *all* problems (hint: errors.Join)
}
```

Then:
- `main` calls `config.Load()` first, and exits with a clear message if it fails.
- `database.InitDB` takes the path as a parameter instead of hardcoding `"test.db"`.
- Add `server/.env.example`, documenting every variable (`.env` itself is already gitignored).
- In `prod`, call `gin.SetMode(gin.ReleaseMode)`.

Use only the standard library (`os.Getenv` / `os.LookupEnv`). No config library needed.

### Why
- **12-factor "config in the environment":** the same binary runs on your laptop, in CI and in Docker, with different config. Code changes should never be needed to change a port.
- **Validate at startup and fail fast.** A typo in `PORT` should crash on boot with a clear message, not cause strange behavior an hour later. `errors.Join` reports *every* problem at once, so you don't fix one, restart and hit the next.
- **A typed struct instead of `os.Getenv` scattered everywhere.** There's one place that knows what config exists, everything else receives plain Go values, and tests can build a `Config{}` directly.
- **`.env.example`** documents the config for other people, including a future reviewer.

### Common mistakes
- Reading env vars deep inside packages (e.g. inside `database`). Only `config` reads the environment; everything else gets values passed in.
- Putting secrets as defaults in code. Phase 2's JWT key will have **no default in prod**, which forces it to be set.

### Check it works
- `PORT=9090 make run` listens on 9090.
- `PORT=abc make run` exits immediately with a readable error.
- Write `config_test.go` using `t.Setenv(...)` for one valid and one invalid case. This is your first test.

---

## Step 4: Structured logging and request IDs
**Branch:** `phase0/logging`

### What to do
1. In `main`, build a logger from config and make it the default:
   ```go
   // dev: human-readable text. prod: JSON (machine-parsable).
   var h slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
   if cfg.Env == "prod" {
   	h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
   }
   logger := slog.New(h)
   slog.SetDefault(logger)
   ```
2. Switch from `gin.Default()` to `gin.New()` and add your own middleware in `internal/middleware/`:
   - **`RequestID()`**:
     - reuse an incoming `X-Request-ID` header if it's a sane length, or generate a new UUID
     - store it with `c.Set("request_id", id)`
     - echo it back in the response header
   - **`AccessLog(logger *slog.Logger)`**: after `c.Next()`, log one line per request with `method`, `path` (use `c.FullPath()`, the route template like `/users/:id`), `status`, `latency_ms`, `request_id` and `client_ip`.
   - **`Recovery`**: keep `gin.Recovery()`, or write your own that logs the panic with the request ID via slog and returns a generic 500.
3. Replace any `fmt.Println` or `log.Println` with `slog`.

### Why
- **Structured logs** (`key=value` or JSON) are *queryable*: "show all 500s for request_id=abc". Plain `Printf` text is only greppable. Every serious log system (Loki, Datadog, CloudWatch) expects structure.
- **`log/slog` is in the standard library** (since Go 1.21). No dependency, and it's the ecosystem standard now.
- **A request ID** ties together every log line produced by one request. When a user reports "it failed at 10:03", you find the one request and see its whole story.
- **`c.FullPath()` instead of the raw URL.** Raw URLs can contain IDs and query strings, and later maybe tokens. Logging the *route template* keeps logs safe and makes them easy to aggregate.
- **The golden rule, starting now:** never log request bodies, passwords, hashes, tokens or `Authorization` headers. Leaked logs are one of the most common real-world breach sources. This matters doubly in a password manager.

### Common mistakes
- Logging `c.Request.URL.String()` (it may contain secrets later) or `c.Request.Header` (it contains `Authorization`).
- Creating a new logger inside every function. Pass one down (or use `slog.Default()` sparingly).
- Trusting a client-supplied `X-Request-ID` without limits. Cap its length (e.g. 64 characters) so nobody can inject huge values into your logs.

### Check it works
`curl -i localhost:8080/users` shows an `X-Request-ID` response header, and the server prints one structured log line containing that same ID.

---

## Step 5: Real `http.Server`, timeouts, graceful shutdown
**Branch:** `phase0/server`

### What to do
Restructure `cmd/server/main.go` into the **`run()` pattern**:

```go
func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. ctx that is cancelled on Ctrl+C / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. config, logger, db, repos, services, handlers, router (from earlier steps)
	//    InitDB should now RETURN an error instead of panicking.

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 3. start the server in a goroutine; send its error on a channel
	// 4. wait for either ctx.Done() (signal) or a server error
	// 5. on signal: srv.Shutdown with a 10s timeout context
	// 6. close the DB (sqlDB, _ := db.DB(); sqlDB.Close())
	// TODO
}
```

Also change `database.InitDB` to return `(*gorm.DB, error)` instead of panicking.

### Why
- **The `run() error` pattern:** `os.Exit` skips `defer`s, so if `main` exits directly, your DB never closes cleanly. With `run()`, every `defer` runs, and there's exactly one exit point.
- **No `panic` for expected failures.** A missing DB file or bad permissions is an *expected* failure, so return an error. Panic is for "this can't happen" programmer bugs.
- **Graceful shutdown:** `Shutdown` stops accepting new connections and *waits* for in-flight requests to finish. Without it, stopping the server (every restart or deploy) can cut a request off halfway, for example mid-write of a vault item.
- **Timeouts:** `http.Server` has **no timeouts by default**. A client can open a connection and send one byte per minute forever (the *Slowloris* attack), exhausting your server. `ReadHeaderTimeout` is the key defense. `gosec` flags a missing one (G112).
- `r.Run()` is a convenience that hides all of this, which is why production code doesn't use it.

### Common mistakes
- Treating `http.ErrServerClosed` from `ListenAndServe` as a failure. It's the *normal* result of `Shutdown`. Check it with `errors.Is`.
- Calling `Shutdown` with `context.Background()`. It could then wait forever on a stuck request. Always give it a deadline.
- On Windows, `SIGTERM` isn't really delivered, but `os.Interrupt` (Ctrl+C) is. Keep both: Docker/Linux uses `SIGTERM`.

### Check it works
Add a temporary slow route (`time.Sleep(5 * time.Second)`), call it, and press Ctrl+C during the request. The request still completes, then the server logs a shutdown and exits with code 0. Remove the slow route before the PR.

---

## Step 6: Domain errors and one error mapper
**Branch:** `phase0/errors`

### What to do
1. Create `internal/apperr/apperr.go`:
   ```go
   package apperr

   // Kinds: categories the HTTP layer knows how to map.
   var (
   	ErrNotFound     = errors.New("not found")
   	ErrConflict     = errors.New("conflict")
   	ErrValidation   = errors.New("validation")
   	ErrUnauthorized = errors.New("unauthorized")
   )

   // Error carries a kind, a stable machine-readable code, a safe message for
   // clients, and the internal cause (never shown to clients).
   type Error struct {
   	Kind    error
   	Code    string // e.g. "user_not_found", "email_taken"
   	Message string // safe to show the client
   	Cause   error  // for logs only
   }

   func (e *Error) Error() string   { /* TODO: include code and cause */ }
   func (e *Error) Unwrap() []error { return []error{e.Kind, e.Cause} } // lets errors.Is see both

   func New(kind error, code, msg string, cause error) *Error { /* TODO */ }
   ```
2. **Repositories** translate GORM errors into *kinds*:
   - `gorm.ErrRecordNotFound` becomes `apperr.ErrNotFound`.
   - A unique-constraint violation becomes `apperr.ErrConflict`. Hint: open GORM with `&gorm.Config{TranslateError: true}` and check `errors.Is(err, gorm.ErrDuplicatedKey)`.
   - Anything else is wrapped with `fmt.Errorf("find user: %w", err)`.
3. **Services** attach meaning. For example, `ErrConflict` from `Save` becomes `apperr.New(apperr.ErrConflict, "email_taken", "email is already registered", err)`.
4. **One mapper**: rewrite `handlers/error.go`:
   ```go
   func respondError(c *gin.Context, err error) {
   	// TODO:
   	// - errors.As(err, &appErr): status from appErr.Kind (errors.Is chain):
   	//     NotFound→404, Conflict→409, Validation→400, Unauthorized→401
   	//   body: {"code": appErr.Code, "message": appErr.Message}
   	// - anything else: 500 {"code":"internal_error","message":"internal server error"}
   	//   and slog.ErrorContext(... "err", err, "request_id", ...) with the FULL error
   }
   ```
5. Update `models.ErrorResponse` to `{code string, message string}`. `code` is now a **string**, not the HTTP status (the status is already in the response).
6. JSON binding errors (`ShouldBindJSON`) become `400` with code `invalid_request`. Pick **one** binding method (`ShouldBindJSON`) and use it everywhere.
7. Fix the inconsistency: `UpdateUser` not-found must be **404** (it's currently 400).

### Why
- **Layers shouldn't leak.** Handlers shouldn't import `gorm` (they currently do), and services shouldn't know HTTP status codes. Each layer translates errors into its own vocabulary. That's what lets you swap SQLite for Postgres later without touching a handler.
- **Stable error codes** (`"email_taken"`) are an API contract. Your CLI will `switch` on them. Human messages can change; codes can't.
- **Never send `err.Error()` to clients.** It can contain SQL, table names, file paths or driver versions, which is free reconnaissance for an attacker. Log it with the request ID instead. The client gets a request ID they can report, and you find the full details in logs.
- **Why `Unwrap() []error`?** Since Go 1.20, `errors.Is` walks *multiple* wrapped errors. So `errors.Is(err, apperr.ErrConflict)` and `errors.Is(err, gorm.ErrDuplicatedKey)` both work on the same error value.

### Common mistakes
- Mapping errors in every handler with copy-pasted `if/else` (like now). The point is **one** mapper.
- Comparing with `err == gorm.ErrRecordNotFound`. Always use `errors.Is`, because wrapping breaks `==`. The `errorlint` linter from step 2 catches this.
- Forgetting the 500 path logs the error. A silent 500 is the worst kind to debug.

### Check it works
- `GET /users/99999` returns `404 {"code":"user_not_found",...}`.
- Registering the same email twice returns `409 {"code":"email_taken",...}` with **no SQLite text**.
  If you get a 500 instead, the driver isn't translating the error. Check that `TranslateError: true` is set, and tell me in the PR (it's a good debugging exercise).

---

## Step 7: Interfaces, `context`, first unit tests
**Branch:** `phase0/interfaces-tests`

### What to do
1. **Add `ctx context.Context` as the first parameter** to every service and repository method. In handlers, pass `c.Request.Context()`. In repositories, use `r.db.WithContext(ctx)`.
2. **Define the interface where it's used**, in `internal/services/userservice.go`:
   ```go
   // UserStore is what UserService needs from storage. Defined here (by the consumer),
   // not in repositories.
   type UserStore interface {
   	Create(ctx context.Context, u *models.User) error
   	FindByID(ctx context.Context, id uint) (*models.User, error)
   	FindAll(ctx context.Context) ([]models.User, error)
   	Update(ctx context.Context, u *models.User) error
   	Delete(ctx context.Context, id uint) error
   }

   type UserService struct {
   	store UserStore
   }

   func NewUserService(store UserStore) *UserService { return &UserService{store: store} }
   ```
   `repositories.UserRepository` satisfies it implicitly. Go needs no `implements` keyword. Add constructors (`NewUserRepository(db)`, `NewUserHandler(svc)`) and make fields unexported.
3. **Make every endpoint return DTOs**, including `CreateUser` (it currently returns the whole `models.User`). Add one small helper, `toUserDTO(u *models.User) models.GetUserDTO`, and use it everywhere instead of the three copy-pasted struct literals.
4. **Write `userservice_test.go`** with a hand-written fake:
   ```go
   type fakeStore struct {
   	users  map[uint]*models.User
   	nextID uint
   	err    error // force failures
   }
   // TODO: implement UserStore methods on *fakeStore
   ```
   Then write **table-driven tests**:
   ```go
   func TestUserService_GetUserByID(t *testing.T) {
   	tests := []struct {
   		name     string
   		seed     []*models.User
   		id       uint
   		wantCode string // "" = success
   	}{
   		{name: "found", /* ... */},
   		{name: "not found", /* ... */, wantCode: "user_not_found"},
   	}
   	for _, tt := range tests {
   		t.Run(tt.name, func(t *testing.T) {
   			// TODO
   		})
   	}
   }
   ```
   Cover these at least: register (success, duplicate email), get (found, not found), update (not found), delete (not found).

### Why
- **"Accept interfaces, return structs"** is the core Go idiom. The service depends on *behavior* (`UserStore`), not on GORM. Tests pass a fake, and a future Postgres repository just needs the same methods.
- **Define interfaces at the consumer**, which is the opposite of Java. The service knows which methods it needs, so small interfaces stay small. The repository doesn't need to know who uses it.
- **Hand-written fakes over mocking libraries** (for now). They're plain Go, easy to read and don't break on refactors. You'll learn more than from generated mocks.
- **`context.Context`**: when the client disconnects or a timeout hits, `ctx` is cancelled and GORM aborts the query. Without it, abandoned requests keep burning DB time. It's also how you'll pass request-scoped deadlines later.
- **Table-driven tests** are the Go standard. Adding a case is one line, and `t.Run` gives each case a name in the output.
- **Why test now?** Step 8 changes IDs from `uint` to UUID and rewrites the DB layer. These tests are your safety net.

### Common mistakes
- Putting `UserStore` in the `repositories` package. That's backwards, and it creates an import cycle risk.
- Giant interfaces ("IUserRepository" with 20 methods). Keep only what the consumer calls.
- Storing a `ctx` in a struct field. Always pass it as a parameter.
- Testing the fake instead of the service. Assert on what the **service** returns.

### Check it works
`go test ./... -cover` passes, and `services` coverage is 70% or more.

---

## Step 8: Database done right
**Branch:** `phase0/database`

This is the biggest step. Split it into two PRs if it grows past about 400 lines.

### What to do
1. **SQLite pragmas via the DSN** (glebarez/modernc syntax):
   ```go
   dsn := cfg.DBPath +
   	"?_pragma=foreign_keys(1)" +
   	"&_pragma=journal_mode(WAL)" +
   	"&_pragma=busy_timeout(5000)" +
   	"&_pragma=synchronous(NORMAL)"
   ```
   Add a startup check that runs `PRAGMA foreign_keys;` and fails if it doesn't return `1`.
2. **Versioned migrations with [goose](https://github.com/pressly/goose)**:
   - Put SQL files in `internal/database/migrations/`, e.g. `00001_create_users.sql`:
     ```sql
     -- +goose Up
     CREATE TABLE users (
         id         TEXT PRIMARY KEY,               -- UUID v7
         name       TEXT NOT NULL,
         email      TEXT NOT NULL UNIQUE COLLATE NOCASE,
         created_at DATETIME NOT NULL,
         updated_at DATETIME NOT NULL
     );

     -- +goose Down
     DROP TABLE users;
     ```
   - Embed them in the binary and run them at startup:
     ```go
     //go:embed migrations/*.sql
     var migrationsFS embed.FS

     func Migrate(db *sql.DB) error {
     	goose.SetBaseFS(migrationsFS)
     	// TODO: goose.SetDialect("sqlite3"); goose.Up(db, "migrations")
     }
     ```
   - **Remove `AutoMigrate`.**
   - Add `make migrate-new name=xyz` using the goose CLI to create new migration files.
3. **UUID v7 primary keys** (`github.com/google/uuid`, `uuid.NewV7()`):
   - `User.ID` becomes `uuid.UUID` with `gorm:"type:text;primaryKey"`.
   - The **service** sets `ID` on create, explicitly. Avoid GORM hooks: they're hidden magic.
   - Handlers parse `:id` with `uuid.Parse` instead of `strconv.ParseUint`. An invalid UUID returns 400 `invalid_id`.
   - Update your fake store and tests.
4. **Email normalization**: in the service, use `email = strings.ToLower(strings.TrimSpace(email))` before saving *and* before looking up. The `COLLATE NOCASE` in the schema is a second line of defense.
5. **Handler integration test** in `internal/handlers/userhandler_test.go`:
   - Open a real SQLite DB in `t.TempDir()`, run migrations, and wire the real repository, service, handler and router.
   - Use `httptest.NewRecorder()` plus `router.ServeHTTP(...)`.
   - Test: `POST /users` → 201; the same email in different case → 409 `email_taken`; `GET /users/not-a-uuid` → 400.

### Why
- **`foreign_keys(1)`**: SQLite **ignores** foreign keys by default (for backward compatibility). Without it, `ON DELETE CASCADE` silently does nothing, so deleting a user leaves their encrypted vault behind. Your startup check guarantees it's on, because trusting config you haven't verified is how security bugs happen.
- **WAL mode** lets readers work while a writer writes (the default mode locks the whole DB). **`busy_timeout`** makes a blocked writer wait up to 5 seconds instead of instantly failing with `database is locked`. Together they make SQLite behave well under a concurrent HTTP server.
- **Migrations instead of `AutoMigrate`:**
  - `AutoMigrate` never drops or renames columns, can't do data migrations, and you can't review what it will do.
  - Migration files are code-reviewed SQL, applied in order and recorded in a `goose_db_version` table, so every environment has exactly the same schema history.
  - Embedding them means the binary carries its own schema, so there are no missing-file deploy bugs.
- **UUIDs:**
  - Sequential IDs leak your user count (`/users/1042` means about 1042 users) and invite enumeration (`/users/1`, `/2`, …).
  - UUIDs are unguessable and can be generated without asking the DB.
  - **v7** is time-ordered, so new rows append to the end of the index instead of landing at random positions (better write performance than v4).
- **Email normalization:** without it, `Bob@x.com` and `bob@x.com` are two accounts. That's confusing at best, and an account-takeover trick at worst. Normalize in one place (the service) so lookups and inserts always agree.
- **The integration test** checks what unit tests can't: that your SQL, pragmas, constraints and HTTP wiring actually work together. `t.TempDir()` gives each test a fresh, automatically deleted DB (no shared state between tests).

### Common mistakes
- Using `:memory:` with a connection pool. Each pooled connection gets its *own* empty database, which causes flaky "no such table" errors. A temp file avoids this.
- Editing a migration that has already run. **Never edit applied migrations**; add a new one. (Your data is disposable right now, but build the habit.)
- Forgetting `COLLATE NOCASE` or normalization on *lookup*, so the insert is normalized but `FindByEmail` isn't.
- Letting GORM create the table anyway because some `AutoMigrate` call survived. Search for it and remove it.

### Check it works
- Delete the old DB file and start the server. The logs show the migration applied.
- `sqlite3 passwordmanager.db "PRAGMA journal_mode;"` prints `wal`.
- `POST /users` returns a UUID `id`.
- `make test lint` is green, and CI is green.

---

## Phase 0: Definition of Done

- [ ] `make lint test` green locally; CI green on `main`
- [ ] Standard layout (`cmd/server`, `internal/…`); no dead code; `go.mod` tidy
- [ ] Config from env with validation; `.env.example` exists
- [ ] Structured logs with request IDs; no secrets logged
- [ ] Graceful shutdown works; server timeouts set; no `panic` for expected errors
- [ ] One error mapper; stable error codes; no internal error text in responses
- [ ] Services depend on interfaces; `ctx` everywhere; table-driven unit tests
- [ ] Foreign keys verified on; WAL; goose migrations; UUID v7 IDs; normalized emails
- [ ] A duplicate email (any case) → `409 {"code":"email_taken"}`

When all boxes are ticked, tell me. I'll do a whole-phase review, then write `docs/guides/phase-1.md`.

## Questions to think about (answer in your PRs or ask me)
These don't block you. They're here to build judgment.
1. In step 6, why is it OK for repositories to import `apperr`, but not for services to import `gorm`?
2. In step 7, what would you lose if the service accepted `*repositories.UserRepository` instead of `UserStore`?
3. In step 8, what could go wrong if two server instances ran migrations at the same moment? (Hint: this is one reason bigger systems run migrations as a separate deploy step.)
