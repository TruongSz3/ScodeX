# ADR-009: Multi-Agent Team Lifecycle

- Status: accepted
- Date: 2026-03-05

## Context
Need first-class multi-agent execution for parallel and chained tasks.

## Decision
Adopt explicit team lifecycle primitives:
- `spawn`
- `send`
- `resume`
- `wait`
- `close`

with depth/thread/runtime limits and auditable team events.

## Consequences
- Pros: deterministic team orchestration
- Cons: higher coordination complexity

## Alternatives
- Single-agent-only runtime
