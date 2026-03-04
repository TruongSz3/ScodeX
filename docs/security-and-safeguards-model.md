# Security and Safeguards Model

## Safety Layers
1. Rules Engine (`allow|ask`)
2. Approval lifecycle
3. Runtime Safeguards (`allow|deny`, final gate)

## Policy Modes
- `ask` (default)
- `auto_allow_all` (no approval prompts; safeguards still apply)

## Runtime Safeguard Examples
- outside workspace path
- malformed command segmentation
- missing approval token for ask-gated action
- mid-stream provider account switch
- invalid patch structure
- untrusted provider endpoint

## Credential and Data Handling
- Provider keys in secure local store when possible
- Redact sensitive values in logs/events
- Keep provider cache/thread metadata out of persistence

## Network Controls
- Loopback bind by default
- Optional allowlist/deny controls
- Explicit deny precedence in network policy evaluation

## Security Invariants
- Safeguards are non-overridable
- Imported/custom definitions cannot bypass global security caps
- Every rejection emits auditable event and reason code
