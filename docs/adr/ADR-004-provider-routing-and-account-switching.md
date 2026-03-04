# ADR-004: Provider Routing and Account Switching

- Status: accepted
- Date: 2026-03-05

## Context
Need provider-only inference with multi-account reliability.

## Decision
Support two in-provider switching modes:
- `turn_robin`
- `quota_failover` (default)

No cross-provider execution failover in v1.

## Consequences
- Pros: reliability control with clear behavior
- Cons: added scheduler complexity

## Alternatives
- Single account only
- Cross-provider failover in MVP
