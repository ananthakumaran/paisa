# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Paisa is a personal finance manager that wraps the external `ledger` (or `hledger` / `beancount`) plain-text accounting CLI with a Go backend and a SvelteKit web UI. Two distributions share the same backend: a server binary (`paisa serve`) and a Wails desktop app (under `desktop/`).

## Common commands

Development:
- `make develop` — runs Go server (with nodemon hot reload) and Vite dev server concurrently. Vite proxies `/api/*` to `:7500`.
- `make debug` — same as develop but freezes "now" to 2022-02-07 so the fixture data renders deterministically.
- `make docs` — serves the MkDocs site at `:8000`.
- `cd desktop && make develop` — runs the Wails desktop app in dev mode (needs `wails` CLI, tag `webkit2_40` on Linux).

Build:
- `make jsbuild` (or `npm run build`) — SvelteKit static build writes to `web/static/`, which is then `go:embed`-ed by `web/web.go`. The Go binary must be rebuilt after the frontend.
- `make install` — full build (`npm run build && go build && go install`).
- `make windows` — cross-compile to Windows (requires mingw).

Lint / format:
- `make lint` — runs `prettier --check src`, `npm run check` (svelte-check), and `gofmt -l .` (must be empty).
- `npm run format` — apply Prettier.

Tests:
- `make test` — full suite: builds JS, runs `bun test --preload ./src/happydom.ts src`, runs regression tests in `tests/` (these start a real `paisa serve` and diff API responses against JSON fixtures), then `go test ./...`.
- `make jstest` — just the JS / regression tests.
- `make regen` — regenerate the regression fixtures (`REGENERATE=true bun test tests`). Use after intentional API changes.
- Single Go test: `go test ./internal/xirr/...` or `go test -run TestName ./internal/...`.
- Single JS test: `bun test --preload ./src/happydom.ts src/lib/journal.test.ts`.

Parser regeneration (Lezer grammars for the sheet DSL and search query language):
- `npm run parser-build` — regenerates `src/lib/sheet/parser.js` and `src/lib/search/parser/parser.js` from the `.grammar` files. Run this whenever a grammar changes; the generated files are committed.

## Architecture

### Backend (Go)

Entry points:
- `paisa.go` → `cmd/` — CLI surface via Cobra. `serve` (`cmd/serve.go`) is the web server, `update` syncs data, `init` writes a sample config.
- `desktop/main.go` — Wails app that mounts the same `server.Build(...)` HTTP handler as the in-process asset server (no separate API binary).

`cmd/root.go` resolves config in this order: `--config`, `$PAISA_CONFIG`, `./paisa.yaml`, then `$XDG_DOCUMENTS_DIR/paisa/paisa.yaml` (auto-generates a minimal one if missing). Config is only loaded for `serve` and `update`.

`internal/` is the business core:
- `internal/server` — Gin router. All routes are declared in `server.go::Build()`; one file per logical area (`expense.go`, `networth.go`, `portfolio.go`, …). `TokenAuthMiddleware` enforces basic-auth + rate limiting only when `user_accounts` is set in config.
- `internal/ledger` — shells out to `ledger` / `hledger` / `beancount` (binary chosen by config). `ledger.Cli()` returns the active implementation. Parsing is done by running the CLI and reading its output, not by reimplementing the format.
- `internal/model` — GORM models + sync orchestration. `SyncJournal` validates the journal via the CLI, then upserts `posting.Posting` and `price.Price`. `SyncCommodities` / `SyncPortfolios` / `SyncCII` pull from scrapers.
- `internal/scraper` — provider implementations grouped by asset class (`mutualfund`, `nps`, `stock`, `metal`, `india`). `GetProviderByCode` resolves a provider by string code from config.
- `internal/accounting`, `internal/query`, `internal/taxation`, `internal/xirr`, `internal/prediction` — computation helpers used by server handlers.
- `internal/cache` — process-wide cache; `Sync` clears it on every refresh.
- `internal/config` — YAML config + JSON schema (`schema.json` is also served to the frontend so the editor can validate).
- `internal/binary` — locates the external `ledger`/`hledger`/`beancount` binary, falling back to a bundled one for desktop builds.

Database: SQLite via GORM, opened by `utils.OpenDB()` (path under XDG state). `model.AutoMigrate` runs on every `serve` / `update`.

### Frontend (SvelteKit, static export)

- `src/routes/(app)/...` — pages. `+layout.ts` disables SSR/prerender; the app is a pure SPA. SvelteKit's static adapter writes to `web/static`, which `web/web.go` embeds (`//go:embed all:static`). The Go server serves `index.html` for any non-`/api`, non-`/_app/*` path (see `server.go::NoRoute`).
- `src/lib/<feature>.ts` mirrors backend handlers (`expense.ts`, `investment.ts`, …) — each is the typed fetch + view-model layer for one Go handler. Components in `src/lib/components/` are mostly presentational.
- `src/lib/sheet/` — custom spreadsheet DSL: Lezer grammar (`language.grammar`) compiled to `parser.js`, evaluated by `interpreter.ts`. Built-in functions in `functions.ts`. The sheet UI lives in `src/routes/(app)/more/sheets`.
- `src/lib/search/parser/` — Lezer grammar for the transaction search query language, consumed by `src/lib/search_query_editor.ts`.
- `src/lib/editor/`, `src/lib/template_editor.ts`, `src/lib/handlebars_parser.ts` — CodeMirror-based ledger file editor and Handlebars-based import templates (CSV / PDF / XLSX → ledger entries).

### Tests

- `tests/regression.test.ts` — end-to-end regression. For each fixture directory in `tests/fixture/`, it launches `paisa serve -p 5700`, hits every `/api/*` endpoint, and diffs the response against the committed JSON snapshot. Diffs cause failures; `REGENERATE=true` overwrites snapshots. This requires the `paisa` binary to already be built (`go build`) and the `ledger` CLI to be on PATH.
- `internal/.../...test.go` — standard Go unit tests, mainly around `ledger` parsing, `xirr`, and `utils`.
- JS unit tests live next to source as `*.test.ts` (run under `bun:test` with `happydom`).

## Conventions worth knowing

- Decimals: backend uses `shopspring/decimal` with `MarshalJSONWithoutQuotes = true`. Frontend uses `bignumber.js`. Don't introduce native floats for money.
- `utils.SetNow` / `utils.Now()` — all date-sensitive code uses this rather than `time.Now()` so tests can pin "today" via `--now` or `TZ=UTC`.
- When adding an API: register the route in `internal/server/server.go` and add a typed fetcher in `src/lib/<feature>.ts`. Regenerate fixtures with `make regen` afterwards.
- Readonly mode: when `config.GetConfig().Readonly` is true, mutating endpoints must short-circuit with `{success: true}` (see existing handlers — this is how the demo deployment is locked down).
- Prettier: `printWidth: 100`, `trailingComma: "none"`. Go: standard `gofmt`. CI lint will fail on any `gofmt -l` output.
