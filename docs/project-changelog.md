# Project Changelog

## 2026-03-05

### Runtime delivery snapshot
- Implemented Go runtime skeleton with `agentd` daemon and `agent` CLI binaries.
- Implemented handshake endpoints with `X-Agent-Token` auth and `sessionId` match enforcement.
- Added local HTTP fallback daemon on `127.0.0.1:7777` with non-loopback opt-in guard.
- Added in-memory event bus and session store.
- Added SQLite bootstrap/migrations, FTS5 lexical index schema/repository, indexing pipeline, and rerank flag default OFF.

Status note:
- Phase 1: completed
- Phase 2: in-progress (~75%)

### Added
- Initialized `docs/` baseline documentation set.
- Added architecture summary docs aligned with blueprint v2.3.
- Added Go tech stack documentation baseline.
- Added contract docs (`Agent=WHO`, `Skill=HOW`, API contract, safeguards model, implementation checklists).
- Added ADR set (`ADR-001` to `ADR-011`) under `docs/adr/`.

### Updated
- Runtime blueprint advanced to include audited consistency fixes and Go stack decisions.
- Progress sync: plan `plans/20260305-go-mvp-stack-for-local-first-agent/plan.md` moved `pending -> in-progress`; Phase 1 marked complete, Phase 2 marked in-progress with evidence-mapped paths.
- Roadmap pointers synced for consistency in `docs/project-roadmap.md` and `docs/development-roadmap.md`.

### Notes
- Superseded import-pipeline planning artifacts were cancelled previously.
