# ADR-001: Local Daemon + Multi-Client Architecture

- Status: accepted
- Date: 2026-03-05

## Context
Need one runtime serving both CLI/TUI and local web UI without duplicating core logic.

## Decision
Use a single local daemon (`agentd`) with multiple clients (`agent` CLI/TUI and web UI).

## Consequences
- Pros: shared orchestration, policy, storage, event bus
- Cons: process lifecycle and IPC complexity

## Alternatives
- CLI monolith only
- Desktop all-in-one app
