# ADR-007: Control Layer Contracts

- Status: accepted
- Date: 2026-03-05

## Context
Need full customization across runtime behavior surfaces.

## Decision
Define first-class registries for:
- agents
- skills
- workflows
- commands/slash commands
- tools
- plugins
- hooks
- team orchestration

## Consequences
- Pros: extensibility and consistent governance
- Cons: larger schema and lifecycle surface

## Alternatives
- Hard-coded behavior only
- Partial registry model
