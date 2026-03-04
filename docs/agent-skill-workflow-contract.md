# Agent-Skill-Workflow Contract

## Core Rule
- Agent = WHO
- Skill = HOW

## Agent (WHO)
Agent definitions own:
- identity and role
- authority boundary
- delegation limits
- default runtime preferences

Agent definitions must not contain:
- procedural step-by-step playbooks
- implementation checklists

## Skill (HOW)
Skill definitions own:
- procedural method
- reusable execution guidance
- tool usage pattern

Skill definitions must not contain:
- ownership authority
- escalation rights

## Workflow Step Binding
Every step must bind:
- `role_agent_id`
- `skill_id`
- `tool_scope`

Missing any binding = invalid workflow step.

## Required Validators
- `role_skill_boundary_lint`
- `workflow_binding_lint`
- `capability_lint`

## Runtime Enforcement
- Validation must pass before activation
- Runtime safeguards can still deny execution if unsafe
