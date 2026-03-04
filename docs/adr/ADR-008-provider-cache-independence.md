# ADR-008: Provider Cache Independence

- Status: accepted
- Date: 2026-03-05

## Context
Multi-account routing can break provider-side cache/thread affinity.

## Decision
Treat local session/context as source of truth and ignore provider cache/thread affinity.

## Consequences
- Pros: deterministic behavior across account switches
- Cons: potentially higher provider token usage

## Alternatives
- Depend on provider-side cache/thread semantics
