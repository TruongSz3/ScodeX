# Code Standards

## Core Principles
- YAGNI, KISS, DRY
- Safety and determinism over feature breadth
- Explicit contracts over implicit behavior

## Language and Stack
- Primary implementation language: Go
- Runtime binaries: `agentd`, `agent`
- Storage: SQLite (WAL)

## Go Standards
- Use clear package boundaries under `internal/`
- Prefer explicit interfaces at component boundaries
- Thread cancellation through `context.Context`
- Return typed errors and wrap root causes
- Use structured logging for all state-changing operations

## Runtime Contract Rules
- Agent manifests define WHO only (authority/ownership/delegation)
- Skill manifests define HOW only (procedures/methods)
- Workflow steps must bind WHO + HOW + tool scope
- Registry validation is required before activation

## API and Event Standards
- Mutating APIs require `runtime.initialize` handshake first
- Every side-effect emits auditable events
- Keep method and event names stable and versioned

## Tool and Policy Standards
- Rules Engine returns only `allow|ask`
- Runtime Safeguards are final `allow|deny` gate
- Canonical tool IDs must be used in rules and execution traces

## Testing Standards
- Unit tests for router/policy/validators
- Integration tests for approval + safeguards + tool execution
- E2E tests for workflow and team orchestration
- Contract tests for WHO/HOW boundary lints
