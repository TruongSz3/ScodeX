# ADR-002: Tool Runtime Safety and Policy Model

- Status: accepted
- Date: 2026-03-05

## Context
Need execution safety without blocking normal developer workflows.

## Decision
Adopt two-plane model:
- Policy Engine: `allow|ask`
- Runtime Safeguards: `allow|deny` final gate

## Consequences
- Pros: clear user-facing approvals + hard technical safety boundary
- Cons: more moving parts and testing overhead

## Alternatives
- Policy-only deny/allow without safeguard plane
- Fully permissive runtime mode without hard denials
