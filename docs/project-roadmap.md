# Project Roadmap

This file mirrors the active engineering roadmap and is intended as the canonical product-facing roadmap document.

## Phase 0 - Scope Lock
- Freeze contracts: `Agent=WHO`, `Skill=HOW`
- Freeze provider switching modes and safety defaults

## Phase 1 - Runtime Skeleton
- Status: completed (sync 2026-03-05)
- `agentd` daemon + `agent` CLI/TUI bootstrap
- Session lifecycle and event bus

## Phase 2 - Context Engine
- Status: in-progress (~75%, sync 2026-03-05)
- Incremental indexing
- Lexical retrieval baseline
- Optional semantic rerank flag

## Phase 3 - Orchestration and Safety
- Orchestrator loop
- Policy engine (`allow|ask`)
- Runtime safeguards (`allow|deny`)

## Phase 4 - Control Layer
- Agents/Skills/Workflows
- Commands/Tools/Plugins/Hooks
- WHO/HOW binding validators

## Phase 5 - Local Web UI
- Session timeline + approvals
- Patch view/apply
- Team execution view

## Phase 6 - Reliability and Release
- Provider scheduler hardening
- Resiliency tests
- Cross-platform packaging and rollback
