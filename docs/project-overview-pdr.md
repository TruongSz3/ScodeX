# Project Overview (PDR)

## Project
- Name: Local-First AI Coding Agent Runtime
- Audience: solo developers
- Primary surfaces: terminal CLI/TUI and local web UI

## Problem
Developers need a coding agent that can act on real repositories with strong local control, explicit approvals, and auditable actions. Cloud-only assistants often fail privacy and reproducibility expectations.

## Product Goals (MVP)
- Chat-to-execution workflow on local machine
- Patch-first code editing with review
- Policy + runtime safeguards for safe tool execution
- Provider-only inference (BYOK/OAuth)
- Multi-account provider switching (`turn_robin`, `quota_failover`)
- Customizable runtime layers: agents, skills, workflows, slash commands, tools, plugins, hooks
- Multi-agent team orchestration primitives

## Core Contracts
- Agent = WHO (identity, ownership, authority)
- Skill = HOW (procedure, method, reusable execution guidance)
- Every workflow step must bind WHO + HOW + tool scope

## Non-Goals (MVP)
- Local on-device model inference
- Cross-provider failover execution
- Multi-IDE parity from day 1

## Success Criteria
- >= 65% task success on curated benchmark
- >= 90% accepted patch ratio for low-risk tasks
- P95 lexical retrieval < 250 ms
- 0 critical policy bypass incidents
- 100% turn-boundary provider account switching correctness

## Current Status
- Architecture/spec phase complete (blueprint v2.3)
- Implementation not started
