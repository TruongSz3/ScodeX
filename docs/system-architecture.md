# System Architecture

## High-Level Topology
```text
[CLI/TUI]      [Local Web UI]
     \            /
      \  local IPC/API
       +----------+
         [agentd]
            |
  +---------+---------+--------------------------------+
  | Orchestrator      | Context Engine                |
  | Provider Router   | Tool Runtime + Safeguards     |
  | Policy Engine     | Control Layer Registries      |
  +-------------------+-------------------------------+
            |
      [SQLite + artifacts + indexes]
            |
   [Provider Connectors (BYOK/OAuth)]
```

## Runtime Planes
- Control plane local by default (sessions, indexing, policy, audit)
- Inference plane external providers only
- Safety plane layered:
  - Rules Engine (`allow|ask`)
  - Approval lifecycle
  - Runtime Safeguards (`allow|deny` final gate)

## Core Components
- Orchestrator: plan -> retrieve -> decide -> act -> verify
- Context Engine: lexical retrieval default, optional semantic rerank
- Tool Runtime: deterministic execution with cwd/timeout/env guards
- Provider Router: per-provider account scheduling (`turn_robin`, `quota_failover`)
- Control Layer: agents, skills, workflows, commands, tools, plugins, hooks
- Team Orchestrator: `spawn/send/resume/wait/close`

## Protocol and Lifecycle
- Active transport: local HTTP daemon (`127.0.0.1:7777` default)
- Non-loopback bind requires explicit opt-in (`AGENTD_ALLOW_NON_LOOPBACK=true`)
- Compatibility profile target: stdio JSONL
- Mandatory handshake before mutating calls:
  - `POST /runtime/initialize` with `protocolVersion="v1"`
  - `POST /runtime/initialized` with matching `sessionId`
- Auth token header enforced on mutating routes: `X-Agent-Token`

## Architectural Contracts
- Agent = WHO
- Skill = HOW
- Workflow step is invalid without:
  - `role_agent_id`
  - `skill_id`
  - `tool_scope`

## Implementation Status (Phases 1-2)
- ✅ Go runtime module initialized (`go.mod`) with binaries:
  - `cmd/agentd` (daemon)
  - `cmd/agent` (CLI)
- ✅ In-memory runtime state implemented:
  - event bus: `internal/events/bus.go`
  - session store: `internal/session/store.go`
- ✅ Handshake and local runtime API implemented:
  - `internal/api/runtime_handler.go`
  - `internal/api/runtime_handshake.go`
- ✅ Local HTTP fallback daemon implemented with loopback guard:
  - `internal/ipc/local_http_server.go`
- ✅ Context engine storage/indexing primitives implemented:
  - SQLite bootstrap + migration runner: `internal/storage/sqlite/bootstrap.go`
  - FTS5 schema migration: `db/migrations/0001_chunks_fts.sql`
  - chunk repository + lexical search: `internal/storage/sqlite/chunk_repository.go`
  - indexing pipeline: `internal/context/indexer/pipeline.go`
  - rerank feature flag default OFF: `internal/context/rerank/feature_flag.go`
