---
title: "Go MVP Stack for Local-First AI Coding Agent"
description: "Pragmatic solo-dev Go stack preserving blueprint decisions and reducing integration risk."
status: in-progress
priority: P1
effort: 8h
branch: unknown
tags: [go, architecture, mvp, daemon, cli, web-ui]
created: 2026-03-05
---

## 0) Progress sync (2026-03-05)

- Sync scope: Phase 1-2 only
- Note: no `phase-XX-*.md` files exist yet; tracking done in this `plan.md` for now

### Phase status snapshot

| Phase | Status | Progress | Evidence (implemented paths) |
|---|---|---:|---|
| Phase 1 - Runtime Skeleton | completed | 100% | `cmd/agentd/main.go`, `cmd/agent/cli.go`, `internal/app/bootstrap.go`, `internal/events/bus.go`, `internal/api/runtime_handler.go`, `internal/ipc/local_http_server.go` |
| Phase 2 - Context Engine | in-progress | 75% | `internal/context/indexer/*`, `internal/context/retrieval/service.go`, `internal/context/rerank/*`, `internal/storage/sqlite/chunk_repository.go`, `db/migrations/0001_chunks_fts.sql` |

### Pragmatic checklist (until phase files exist)

- [x] Phase 1 baseline runtime skeleton landed
- [x] Health/ready + initialize/initialized + session create routes landed
- [x] Local HTTP fallback server + loopback guard landed
- [x] Event bus + session store baseline landed
- [x] Phase 2 indexing pipeline landed
- [x] SQLite chunks + FTS5 schema landed
- [x] Lexical retrieval service landed
- [x] Semantic rerank feature flag scaffold landed (default OFF)
- [ ] Phase 2 semantic rerank execution path wired end-to-end
- [ ] Phase 2 acceptance/perf verification closed

### Roll-up

- Plan overall progress (Phase 1-2 slice): 88%
- Plan overall status: `in-progress` (Phase 2 still open)

## 1) Recommended Go stack by layer (MVP)

- Daemon/runtime core: Go 1.24+, `context`, `errgroup`, `fx` (or minimal manual DI), `gjson/sjson` optional
- IPC transport: primary Unix Domain Socket / Windows Named Pipe via `nhooyr.io/websocket` not needed; use `google.golang.org/grpc` over UDS/pipe with local auth token interceptor
- API contract: gRPC + `connectrpc.com/connect` gateway for browser-friendly HTTP/JSON + streaming
- Event streaming: server-streaming RPC for CLI/TUI; SSE endpoint for web UI
- CLI: `cobra` + `viper` + `lipgloss` (formatting)
- TUI: `charmbracelet/bubbletea` + `bubbles` + `viewport` + `glamour` (diff/markdown rendering)
- Local web serving: daemon serves embedded static SPA via `embed` + `chi` router + SSE
- Storage: SQLite (`modernc.org/sqlite` for CGO-free default) in WAL mode + `sqlc` + `goose`
- Indexing (lexical default): SQLite FTS5 (`bm25`) + tree-sitter bindings (`smacker/go-tree-sitter`) for symbol-aware chunks
- Semantic rerank (optional): pgvector-equivalent not needed in MVP; use `sqlite-vec` only behind feature flag (remote embeddings already optional/off)
- Policy/rules engine: `cel-go` for match expressions + explicit prefix command matcher + precompiled rule trie
- Plugins/hooks: HashiCorp `go-plugin` (RPC-isolated subprocess) + hook bus with timeout/fail-open/closed policy
- Provider adapters: `go-openai` + custom thin HTTP adapter for OpenAI-compatible `base_url + api_key`; OAuth via `golang.org/x/oauth2`
- Observability: `slog` JSON logs + OpenTelemetry SDK (local exporter default) + Prometheus `/metrics` local endpoint
- Packaging/update: `goreleaser` (multi-OS archives/checksums/signing), optional `minisign` signatures, self-update via `go-update`

## 2) Critical layer alternatives (2 each)

### IPC/API
- Alt A: gRPC + Connect (recommended)
  - Pros: typed schema, streaming, one proto source for CLI/TUI/web, stable long-term
  - Cons: proto/tooling overhead for solo dev
- Alt B: JSON-RPC 2.0 over UDS/pipe + SSE
  - Pros: fastest MVP iteration, easy debug with plain JSON
  - Cons: weaker typing/versioning, higher drift risk across clients

### TUI
- Alt A: Bubble Tea stack (recommended)
  - Pros: mature Go TUI ecosystem, clean state-driven updates, good async handling
  - Cons: steeper initial architecture learning
- Alt B: `rivo/tview`
  - Pros: quick widget composition
  - Cons: less ergonomic for streaming/event-driven modern UX, styling constraints

### Web UI serving
- Alt A: Embedded SPA served by daemon (recommended)
  - Pros: single binary deploy path, perfect local-first model, no extra process mgmt
  - Cons: frontend build integration in Go release pipeline
- Alt B: Separate web dev server in MVP
  - Pros: fastest frontend iteration during dev
  - Cons: diverges from product runtime model, harder release UX

### Storage/index
- Alt A: SQLite + FTS5 + optional sqlite-vec (recommended)
  - Pros: one local datastore, transactional event log + retrieval, simple backup
  - Cons: vector tooling in Go still less mature than Postgres ecosystem
- Alt B: SQLite (app state) + Tantivy/Meilisearch sidecar (index)
  - Pros: stronger search tuning/perf at scale
  - Cons: extra process and sync complexity, anti-KISS for solo MVP

### Rules/policy engine
- Alt A: explicit matcher + CEL hybrid (recommended)
  - Pros: keeps prefix/token-boundary guarantees while allowing explainable expressions
  - Cons: dual model complexity if not constrained
- Alt B: OPA/Rego embedded
  - Pros: powerful policy language, policy tooling ecosystem
  - Cons: heavier mental/runtime overhead; overkill for MVP ask/allow

### Plugin runtime
- Alt A: subprocess plugins via `go-plugin` (recommended)
  - Pros: isolation boundary, crash containment, versioned handshake
  - Cons: RPC overhead and version compatibility work
- Alt B: in-process Go plugin (`plugin` stdlib)
  - Pros: low latency
  - Cons: weak cross-platform support, crash/privilege blast radius too high

## 3) Minimal Go project layout

```text
cmd/
  agentd/                 # daemon entrypoint
  agent/                  # cli+tui entrypoint
internal/
  app/                    # wiring/boot lifecycle
  api/                    # connect/grpc handlers + dto mapping
  ipc/                    # uds/pipe listeners, auth interceptors
  orchestrator/           # state machine and step loop
  teams/                  # spawn/send/resume/wait/close
  registries/             # agents/skills/workflows/tools/plugins/hooks
  policy/                 # risk tiering, rule eval, explainability
  safeguards/             # hard deny checks
  runtime/                # tool execution, sandbox, network policy
  provider/               # router, account scheduler, adapters
  context/                # indexer, chunker, retrieval
  memory/                 # session/project memory management
  storage/
    sqlite/               # sqlc queries, migrations, repo impl
  events/                 # event bus + persistence
  web/                    # static serving + sse handlers
  tui/                    # bubbletea models/views/update
pkg/
  contract/               # proto/gen or shared rpc types
  diff/                   # patch helpers
web/                      # frontend source (vite/react or svelte)
db/migrations/
configs/
```

## 4) Build/release approach (cross-platform)

- Toolchain: pinned Go version (`go.mod` + `mise` or `.tool-versions`)
- CGO-free default builds for daemon/cli (`modernc` SQLite) to simplify Win/macOS/Linux distribution
- `goreleaser` matrix: `darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64`
- Artifacts: `agentd`, `agent`, web assets embedded at build (`go:embed`)
- Signing/checksums: SHA256 + optional cosign/minisign
- Upgrade path: `agent self-update` channel + rollback to prior artifact
- Compatibility gate: daemon exposes API version; cli/web enforce semver compatibility matrix at connect time

## 5) Key risks and mitigations

- IPC complexity (UDS + pipe + browser): start with Connect over localhost HTTP+token, add native pipe/UDS in phase 2 while keeping same handlers
- Plugin safety drift: enforce subprocess-only plugins, per-hook timeout, schema-validated IO, deny network by default
- SQLite contention from event stream + index updates: WAL + batched writes + dedicated writer goroutine + periodic checkpoints
- Policy ambiguity (`ask` vs safeguards deny): codify two-plane decision model in code + test vectors from blueprint section 7.7
- Cross-platform shell/runtime variance: command segmentation and allowlist in platform adapters; golden tests per OS shell parser
- Provider adapter entropy: define strict adapter interface + conformance tests against OpenAI-compatible mock server

## 6) Concrete section edits to blueprint

1. Section 9 (API and IPC Blueprint): add decision: "RPC transport = gRPC/Connect; UDS/NamedPipe primary, localhost HTTP fallback for browser and troubleshooting."
2. Section 7.1 (Client Adapters): add: "Daemon emits server-streaming RPC and SSE events from same event bus; CLI/TUI consumes RPC stream; web consumes SSE."
3. Section 7.8 (Storage Layer): specify "SQLite + FTS5 is required lexical engine for v1; vector rerank behind feature flag, default OFF."
4. Section 7.3 (Context Engine): add implementation note "Symbol chunking via tree-sitter parsers; fallback block chunker if parser unavailable."
5. Section 7.12.3 + 7.7 (Rules/Policy): add "Matcher implementation uses token-boundary prefix matcher + CEL constraints; explain output includes matched segment IDs."
6. Section 7.12.7/7.12.8 (Plugin/Hook): lock "Plugins run out-of-process only in MVP; no in-process dynamic loading."
7. Section 13 (Deployment and Packaging): add "GoReleaser-based multi-OS binaries, embedded web assets, and API-version compatibility handshake."
8. Section 10.1 (Agent Team Flow): add "Team operations executed as orchestrator-managed goroutines with bounded worker pool and per-agent context cancellation."

## Unresolved questions

- Should MVP choose proto-first now, or JSON-RPC first then migrate (faster now vs less churn later)?
- Is CGO-free SQLite acceptable for perf targets, or should macOS/Linux use `mattn/go-sqlite3` and keep `modernc` only for Windows?
- Web UI stack preference (React vs Svelte) for solo-dev velocity in this repo?
