# Codebase Summary

## Current Repository State
Core runtime implementation is in progress. Phase 1-2 foundation code exists in Go and is test-backed.

Top-level structure:
- `cmd/`
- `internal/`
- `db/`
- `docs/`
- `plans/`
- `local-first-ai-coding-agent-system-design-blueprint.md`
- `reference/`

## Key Assets
- Runtime entrypoints:
  - `cmd/agentd/main.go`
  - `cmd/agent/main.go`
- Runtime API + handshake:
  - `internal/api/runtime_handler.go`
  - `internal/api/runtime_handshake.go`
- Local HTTP fallback transport:
  - `internal/ipc/local_http_server.go`
- In-memory runtime state:
  - `internal/events/bus.go`
  - `internal/session/store.go`
- SQLite + lexical context engine base:
  - `internal/storage/sqlite/bootstrap.go`
  - `internal/storage/sqlite/chunk_repository.go`
  - `db/migrations/0001_chunks_fts.sql`
- Indexing + retrieval + rerank flag:
  - `internal/context/indexer/pipeline.go`
  - `internal/context/retrieval/service.go`
  - `internal/context/rerank/feature_flag.go`

## Handshake and API Baseline (Implemented)
- Auth header required on mutating routes: `X-Agent-Token`
- `POST /runtime/initialize` requires `protocolVersion`, returns handshake `sessionId`
- `POST /runtime/initialized` requires matching `sessionId`
- `POST /session/create` requires completed handshake
- `GET /health` and `GET /ready` endpoints available

## Implementation Status
- ✅ Phase 1 completed (runtime skeleton)
  - Go module and daemon/CLI binaries
  - handshake endpoints with auth token enforcement
  - local HTTP daemon default bind `127.0.0.1:7777`
  - non-loopback bind opt-in guard
  - in-memory event bus and session store
- 🟡 Phase 2 in-progress (~75%, context engine baseline modules landed)
  - SQLite bootstrap + migration runner
  - FTS5 lexical index schema and repository
  - indexer pipeline for chunk upsert + lexical index rebuild
  - lexical retrieval service scaffolding
  - semantic rerank feature flag default OFF

## Repomix Evidence
- `repomix-output.xml` regenerated on 2026-03-05.
- Compaction includes the implemented Phase 1-2 files (e.g., `cmd/agentd/main.go`, `internal/api/runtime_handler.go`, `internal/context/indexer/pipeline.go`, `db/migrations/0001_chunks_fts.sql`).
