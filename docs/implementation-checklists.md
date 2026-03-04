# Implementation Checklists

## Runtime Foundation
- [x] Initialize Go module layout (`cmd/agentd`, `cmd/agent`, `internal/*`)
- [x] Implement handshake (`POST /runtime/initialize`, `POST /runtime/initialized`)
- [x] Enforce handshake auth token header (`X-Agent-Token`) on mutating runtime routes
- [x] Add local HTTP fallback daemon (`127.0.0.1:7777` default)
- [x] Add non-loopback bind opt-in guard (`AGENTD_ALLOW_NON_LOOPBACK=true`)
- [x] Add in-memory structured event bus
- [x] Add in-memory session store

## Context Engine Foundation
- [x] Add SQLite bootstrap and migration runner
- [x] Add initial migration with FTS5 lexical index (`chunks_fts`)
- [x] Add chunk repository with lexical index rebuild and lexical search
- [x] Add indexer pipeline for file walk/chunk/upsert/index rebuild
- [x] Add lexical retrieval service scaffolding
- [x] Add semantic rerank feature flag (default OFF)

## WHO/HOW Contract
- [ ] Implement agent registry schema
- [ ] Implement skill registry schema
- [ ] Implement workflow step bindings (`role_agent_id`, `skill_id`, `tool_scope`)
- [ ] Add boundary validators (`role_skill_boundary_lint`, `workflow_binding_lint`, `capability_lint`)

## Tool Runtime and Policy
- [ ] Implement canonical tool ID map
- [ ] Implement rules evaluation and explainability
- [ ] Implement approvals lifecycle
- [ ] Implement runtime safeguards and rejection events

## Provider Router
- [ ] Implement `turn_robin` mode
- [ ] Implement `quota_failover` mode
- [ ] Add account health/cooldown tracking
- [ ] Add usage/quota/rate-limit telemetry events

## Team Orchestration
- [ ] Implement `team.spawn/send/resume/wait/close`
- [ ] Enforce depth/thread/runtime caps
- [ ] Add team timeline events

## UI Surfaces
- [ ] TUI approvals + patch view
- [ ] Local web UI timeline + approvals + team view
- [ ] Runtime mode and provider/account status badges

## Quality Gates
- [ ] Unit tests for validators/router/policy
- [ ] Integration tests for approval+safeguards
- [ ] E2E tests for workflow + team
- [ ] Crash/recovery tests for daemon
