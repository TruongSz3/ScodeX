# Local-First AI Coding Agent - System Design Blueprint

Status: draft-v2.3 (audited + Go stack)
Owner: product + engineering
Last updated: 2026-03-05

Locked decisions (2026-03-05):
- Primary user: solo developer
- Primary surfaces: terminal CLI/TUI and local web UI
- Model execution: provider APIs only (no local on-device model inference)
- Custom third-party provider support: `base_url + api_key` (OpenAI-compatible adapter)
- Multi-account per provider: two switching modes
  - `turn_robin` (switch each turn)
  - `quota_failover` (switch on rate-limit/quota/failure)
- No cross-provider execution failover in v1 (single selected provider per session/profile)
- No per-account model divergence in same provider profile (shared model set only)
- Provider-side cache is ignored by design; canonical conversation/cache is local only
- Provider telemetry focus: quota, rate-limit, token usage
- Policy default mode in v1: `ask`
- Optional mode: `auto_allow_all` => auto-allow across low/medium/high risk without approval prompts
- Non-overridable runtime safeguards may still reject unsafe/invalid execution
- Single runtime profile in MVP (`solo-default`) + optional `auto_allow_all` mode
- Baseline strategy: keep Codex-style runtime/safety architecture, adopt OpenCode-style extensibility and customization surfaces
- Tech stack baseline: implement runtime in Go (daemon, IPC/API, orchestration, policy, providers, plugins)
- Agent contract: `Agent = WHO` (role/ownership/authority)
- Skill contract: `Skill = HOW` (procedure/instructions/reusable method)
- Full-custom layers required: agents, skills, workflows, slash commands, tools, plugins, hooks
- Multi-agent teams are first-class (spawn/send/resume/wait/close lifecycle)
- No external kit import pipeline in v1; runtime registries + APIs are source of truth

## 1) Vision and Product Thesis

Build a local-first AI coding agent that can understand a repository, plan changes, execute tools safely, and return reviewable patches with verification signals.

Core thesis:
- Privacy-first by default: execution/logging/indexing stay local; outbound model context is minimized and explicit.
- Reliable execution beats flashy demos: deterministic tooling and safety gates first.
- Provider-first intelligence: BYOK/OAuth multi-provider onboarding with in-provider account rotation.
- Cache independence: no dependency on provider-side prompt/session cache.

## 2) Problem Statement

Developers need an AI assistant that does more than chat:
- Reads and reasons over a real codebase
- Performs controlled edits and commands
- Produces traceable, auditable actions
- Keeps control-plane local while calling external model providers

Traditional cloud-first agents often fail privacy requirements, have weak tool governance, and cannot guarantee reproducibility in local dev environments.

## 3) Goals, Non-Goals, and Success Criteria

### Goals (MVP)
- Chat-to-execution workflow in local machine
- Repo indexing and retrieval (hybrid lexical + vector)
- Patch-first editing with human review
- Tool runtime with permission model and risk tiers
- Session memory and project memory
- CLI/TUI adapter + local web UI adapter
- BYOK + OAuth support for multiple providers
- Custom provider onboarding via `base_url + api_key`
- Multi-account scheduling per provider (`turn_robin` and `quota_failover`)

### Non-Goals (MVP)
- Full autonomous long-running coding without approvals
- Multi-IDE parity on day 1
- Distributed multi-agent orchestration
- Local on-device model inference runtime
- Provider-managed cache/thread affinity features

### Success Criteria (MVP)
- >= 65% task success on curated local coding benchmark
- >= 90% accepted patch ratio for low-risk tasks
- P95 retrieval latency < 250 ms (warm index)
- 0 critical policy bypass incidents in internal testing
- Crash recovery loses <= 1 in-flight step
- >= 99% successful in-provider account switch for simulated timeout/429 scenarios
- 100% turn-boundary account switching correctness (no mid-stream account swap)

## 4) Design Principles

- KISS: start with one orchestrator loop and deterministic tools.
- YAGNI: avoid advanced multi-agent graphs before quality baseline.
- DRY: shared core services for CLI/TUI and web adapters.
- Human-in-the-loop by default for medium/high-risk actions.
- Event-sourced auditability for every state-changing action.

## 5) Architecture Options and Recommendation

### Option A: CLI/TUI-Only Monolith
- Pros: fastest path to usable product for solo dev
- Cons: local web UI added later becomes refactor-heavy

### Option B: CLI/TUI + Local Web UI Adapters over Local Agent Daemon (Recommended)
- Pros: clean separation, reusable core, easy multi-client, safer policy centralization
- Cons: adds IPC and process lifecycle complexity

### Option C: Desktop App All-in-One
- Pros: complete UX control
- Cons: longest time-to-market, larger maintenance burden

Recommendation: Option B.

## 6) High-Level Architecture (C4 - Container View)

```text
[CLI/TUI Client]       [Local Web UI Client]
       |                    |
       +--------- Local IPC (socket/pipe/http+token) --------+
                                                           |
                                                   [Agent Daemon]
                                                           |
    +----------------+---------------+----------------+---------------+------------------+
    |                |               |                   |               |                  |
[Orchestrator] [Context Engine] [Tool Runtime] [Provider Router] [Policy Engine] [Control Layer]
    |                |               |                   |               |                  |
    +----------------+---------------+----------------+---------------+------------------+
                             |
                     [Local Storage Layer]
             (SQLite + vector index + artifacts + logs)
                             |
                 [Provider Connectors (BYOK/OAuth)]
```

## 6.1 Technology Stack (Go)

Runtime baseline:
- Language/runtime: Go 1.24+ (multi-binary: `agentd`, `agent`)
- Concurrency model: goroutines + `context` cancellation

Core stack by layer:
- IPC/API: Connect RPC over Unix domain socket / Windows named pipe, localhost HTTP fallback
- Event transport: server-streaming RPC + SSE bridge for local web UI
- CLI/TUI: `cobra` + `bubbletea` + `bubbles` + `lipgloss`
- Local web UI: static SPA assets embedded via `go:embed`, served by daemon
- Storage: SQLite (WAL) + `sqlc` + versioned migrations
- Retrieval/indexing: SQLite FTS5 lexical index (default), optional vector rerank by feature flag
- Symbol chunking: tree-sitter integration with block-chunk fallback
- Policy/rules: token-boundary prefix matcher + structured constraints engine
- Plugins/hooks: subprocess plugin runtime with per-hook timeout and isolation
- Providers: OpenAI-compatible adapter (`base_url + api_key`) + OAuth (`x/oauth2`)
- Observability: structured logs, metrics, traces
- Packaging/release: cross-platform Go builds with signed artifacts and rollback-capable updater

## 7) Core Components Blueprint

## 7.1 Client Adapters (CLI/TUI + Local Web UI)

Responsibilities:
- Capture user intent, display plan/patch/approvals
- Stream agent events and tool outputs
- Provide diff review UX and apply/reject actions

Interfaces:
- `session.create`, `session.resume`, `message.send`
- `approval.resolve`
- `patch.preview`, `patch.apply`

MVP scope:
- Terminal TUI with approval overlay + diff actions
- Local web UI panel with session timeline + permission dock
- Local web UI is served by daemon as static assets + local API/SSE stream

Brainstorm prompts:
- Chat-first or plan-first default UX?
- TUI-first vs web-first default launch mode?
- How much command output should be auto-collapsed?

### Reference-Informed Patterns (Codex + OpenCode)

Evidence-derived patterns to adopt:
- Server-first local architecture: one daemon, many clients (CLI/TUI + web).
- Approval and permission as first-class workflow: ask/allow with risk-tiered prompts.
- Safety in layers, not one switch: command policy + cwd confinement + sandbox mode + optional network gate.
- Session lifecycle as product core: resumable sessions, summaries/compaction, traceable actions.
- Local web security baseline: bind localhost by default; if LAN exposed then require auth and strict CORS.
- Codex-style strict safety planes (exec policy + sandbox + network + process hardening).
- OpenCode-style extensibility planes (agent/skill/command/tool/plugin/hook registries).
- Contract discipline: Agent decides WHO acts, Skill defines HOW work is done.

Do not copy blindly:
- Avoid websocket-only control path if marked experimental in upstream patterns.
- Avoid enterprise-heavy control-plane scope in MVP for solo developer target.

## 7.12 Extensibility and Control Layer

Purpose:
- Keep core orchestrator simple while enabling fully customizable behavior.
- Provide explicit contracts for agents, skills, workflows, slash commands, tools, plugins, hooks, and team orchestration.
- Follow contract: `Agent = WHO`, `Skill = HOW`.

MVP boundary:
- Local-first, deterministic, auditable.
- No public plugin marketplace in MVP.
- No dynamic code execution from untrusted remote sources.

### 7.12.0 Agents Registry (WHO)

Responsibilities:
- Define execution identity/ownership (who is acting).
- Define authority boundaries (tool scope, escalation rights, delegation limits).
- Define runtime defaults per role (provider/model, safety mode, step budget).

Agent manifest (customizable):
- `id`, `name`, `description`, `mode` (`primary|subagent|all`)
- `system_prompt_ref`
- `tool_allowlist[]`
- `permission_defaults`
- `provider_preferences`
- `delegation_limits` (`max_depth`, `max_threads`, `max_runtime_seconds`)

Agent resolution order:
- workspace custom agents > user custom agents > built-in agents

### 7.12.1 Skills Registry (HOW)

Responsibilities:
- Register and load reusable capabilities (planning, debugging, testing, docs, etc.)
- Bind each skill to declared permissions and tools
- Keep skills procedural and role-agnostic (method, not ownership)

Minimal manifest:
- `id`, `version`, `description`
- `entrypoint` (local path/module)
- `required_tools[]`
- `required_permissions[]`
- `input_schema`, `output_schema`

MVP policy:
- Built-in and local trusted skills only
- Skill activation requires permission check
- Skill execution emits full audit event trail
- Skill sources are customizable: workspace paths, user paths, optional trusted URLs

### 7.12.2 Runtime Profile and Mode

Responsibilities:
- Keep one explicit runtime profile for solo developer workflow
- Expose one lightweight behavior toggle for dangerous actions

Profile fields (MVP):
- `name` (fixed: `solo-default`)
- `provider_model_preferences` (primary provider, allowed model set)
- `tool_allowlist`
- `step_budget`
- `context_budget_tokens`
- `dangerous_action_mode` (`ask` default | `auto_allow_all` optional)
- `mode_scope` (`session` default | `workspace` persisted)

Config precedence:
- session override > workspace profile > user profile > built-in defaults

Mode persistence rule:
- `session` scope resets on session end/resume of a new session
- `workspace` scope persists for the current workspace until changed

### 7.12.3 Rules Engine

Responsibilities:
- Evaluate action requests and return explainable decision
- Resolve conflicts among system/workspace/session rules

Rule model:
- Match fields: `actor`, `tool`, `command_prefix`, `path_pattern`, `risk`, `network`
- Decision: `allow | ask`
- Optional constraints: `timeout_ms`, `cwd_scope`, `max_output_bytes`

Rule language style (Codex-inspired, MVP):
- Prefix-oriented command rules with explicit token boundaries
- Narrow rule scopes preferred over wildcard-heavy rules
- Rule-level justification is mandatory for auditability

Precedence (MVP):
- Phase 1 (base rules): `ask` wins over `allow` when conflicts exist
- Phase 2 (session grant): explicit `accept_once`/`accept_for_session` may override base `ask`
- Runtime safeguards are final and cannot be overridden

Load-time validation:
- Ruleset includes optional inline test vectors: `match[]`, `not_match[]`
- Ruleset failing validation cannot be activated
- Command string is segmented before evaluation (`|`, `&&`, `||`, `;`, subshell boundaries)
- Each segment is evaluated independently; aggregate result uses strictest decision (`ask > allow`)

Explainability requirement:
- Every decision returns `{rule_id, reason, effective_constraints}`

### 7.12.4 Workflow Registry

Responsibilities:
- Provide reusable execution flows instead of ad-hoc agent behavior

Template shape:
- `workflow_id`, `version`, `steps[]`, `entry_conditions`, `exit_conditions`
- Each step declares agent role (WHO), skill (HOW), tool scope, and approval behavior

MVP templates:
- `plan`
- `implement-safe`
- `fix-regression`
- `test-and-verify`
- `review-and-summarize`

MVP constraints:
- Linear flow only (no DAG scheduler)
- Resume from failed step supported
- Team fan-out steps allowed through explicit team-orchestrator operations

### 7.12.5 Slash Command Registry

Responsibilities:
- Parse slash commands and route to workflow/template or direct control operation

MVP command set:
- `/plan`
- `/fix`
- `/test`
- `/review`
- `/agent`
- `/agents`
- `/skill`
- `/skills`
- `/workflow`
- `/command`
- `/tool`
- `/plugin`
- `/hook`
- `/team`
- `/mode`
- `/policy`
- `/status`

Routing contract:
- `slash_command` -> `action_type` (`workflow|agent|skill|tool|plugin|team|runtime`) + `target` + `args`
- Unknown command returns suggestions, not silent failure

### 7.12.6 Tool Registry

Responsibilities:
- Register built-in, custom, MCP, and plugin tools under one normalized schema.
- Enforce tool contracts and permission gates before execution.

Tool registration sources:
- Built-in tool set (canonical IDs):
  - `read`, `grep`, `glob`
  - `edit_patch`, `edit_write`
  - `shell`, `git`, `test`
  - `task`, `skill`
- Workspace custom tools
- Plugin-contributed tools
- MCP discovered tools

Tool contract:
- `name`, `description`, `input_schema`, `output_schema`, `risk_level`, `executor`
- `aliases[]` optional for legacy compatibility (`patch_apply -> edit_patch`, `write_file_direct -> edit_write`, `search -> grep`)

### 7.12.7 Plugin Runtime

Responsibilities:
- Load trusted plugins and expose plugin capabilities.
- Support plugin-level customizations for tools, commands, hooks, auth, and settings.

Plugin manifest (MVP):
- `name`, `version`, `source`, `enabled`
- `commands[]`, `tools[]`, `hooks[]`, `skills_paths[]`, `mcp_servers[]`

### 7.12.8 Hook Bus

Responsibilities:
- Allow controlled plugin extensions around key lifecycle points
- Keep side effects bounded and observable

MVP hook points:
- `before_model_call`
- `before_tool_call`
- `after_tool_call`
- `before_patch_apply`
- `on_session_complete`
- `permission_ask`
- `command_execute_before`
- `tool_definition`

Execution policy:
- Per-hook timeout
- Fail-closed for security-sensitive hooks, fail-open for non-critical UI hooks
- Hook errors are logged and surfaced in diagnostics

Security constraints:
- Hooks run with least privilege
- No implicit network access unless explicitly granted
- Hook output validated against schema

### 7.12.9 Agent Team Orchestrator

Responsibilities:
- Coordinate multiple sub-agents in parallel and sequential chains.
- Enforce team limits and isolation boundaries.

Core team operations:
- `team.spawn(agent, task)`
- `team.send(team_member_id, input)`
- `team.resume(team_member_id)`
- `team.wait(team_run_id | team_member_id)`
- `team.close(team_member_id)`

Team safety controls:
- Max depth, max active threads, runtime budget per sub-agent
- File ownership hints to reduce overlap conflicts
- Team events are streamed and auditable per agent thread

### 7.12.10 Customization Surface (All Layers)

Customization sources (MVP):
- Workspace config files (project-local)
- User-global config files
- Local manifest directories (`agents/`, `skills/`, `commands/`, `tools/`, `plugins/`)
- RPC-based runtime updates (for UI/CLI driven customization)

Customization guarantees:
- Agents are customizable (WHO): role metadata, permissions, delegation limits, model preferences
- Skills are customizable (HOW): instructions, scripts, references, tool requirements
- Workflows are customizable: template steps, entry/exit conditions, team fan-out steps
- Slash commands are customizable: alias mapping, command templates, argument schemas
- Tools are customizable: register/unregister custom tools and schema contracts
- Plugins are customizable: install/enable/disable, manifest-defined capabilities
- Hooks are customizable: enable/disable, timeout, fail-open/closed behavior
- Agent Teams are customizable: spawn policy, depth/thread caps, coordination strategy

## 7.2 Orchestrator

Responsibilities:
- Execute finite loop: Plan -> Retrieve -> Decide -> Act -> Verify -> Summarize
- Track step budget and stopping conditions
- Handle retries and escalation to user

State machine:
- `idle`
- `planning`
- `awaiting_approval`
- `delegating`
- `executing`
- `aggregating`
- `verifying`
- `completed`
- `failed`

Guardrails:
- Max steps per task (default 10)
- Max consecutive tool failures (default 2)
- Max active team agents (default 3)
- Ask user when confidence < threshold or conflicting evidence

Brainstorm prompts:
- Should plan be explicit by default?
- What is ideal retry policy per tool category?
- Should orchestrator auto-run tests for every edit task?

## 7.3 Context Engine (Indexing + Retrieval)

Responsibilities:
- Parse repository structure and symbols
- Build searchable index (lexical + semantic)
- Provide ranked context bundles per step

MVP design:
- File watcher incremental indexing
- Chunking by symbol/function/class and fallback by block
- Hybrid ranking:
  - lexical score (BM25)
  - embedding similarity
  - path priors (proximity to edited files)

Key constraints:
- Branch switch invalidation support
- Ignore generated/vendor directories by policy

Context Engine defaults (v1):
- Index scope:
  - Track all repository-tracked text files by default
  - Untracked files are indexed only when opened/referenced in session
  - Ignore source: `.gitignore` + built-in denylist (`.git`, `node_modules`, `dist`, `build`, `.cache`, binaries)
- Chunking:
  - Primary: symbol-aware chunks (function/class/module boundaries)
  - Fallback: block chunks with overlap for non-parseable files
  - Every chunk stores stable file path + line ranges + content hash
- Retrieval pipeline:
  1. Query normalization (intent terms + path hints)
  2. Lexical shortlist using BM25 (`top_n = 40`)
  3. Optional semantic rerank (`top_40 -> top_12`) when enabled
  4. Path proximity and recency boost
  5. Diversity cap (`max 4 chunks per file`) to prevent single-file domination
  6. Context pack by token budget with source citation metadata

Locked decisions for v1:
- Embedding rerank default: `OFF` (lexical-first baseline)
- Remote embedding toggle: `ON` only with explicit workspace consent
- If remote embedding is enabled, outbound text is minimized and redacted before request
- Branch handling: workspace-global index with lazy reindex via file hash invalidation on branch switch
- Provider timeout fallback: semantic rerank auto-disabled for current turn, lexical path continues

Context budget policy:
- Reserve token budgets by layer: user request, plan state, retrieval bundle, tool traces
- Retrieval bundle target: top 6-12 chunks depending on remaining budget
- Hard cap prevents context overflow; excess chunks are summarized before inclusion

Quality and latency targets:
- P95 lexical retrieval: < 250 ms (warm index)
- P95 hybrid retrieval: < 600 ms (when semantic enabled)
- Recall@10 on internal benchmark: >= 0.75

## 7.4 Tool Runtime

Responsibilities:
- Execute tool calls safely and reproducibly
- Standardize outputs for orchestrator
- Enforce timeout, cwd, env policies

Tool categories:
- File tools: read, search, patch apply
- Git tools: status, diff, add/commit (if enabled)
- Shell tools: build/test/lint scoped commands
- Test tools: predefined test runners

MVP safety:
- Command allowlist + runtime hard-block safeguards
- Working directory confinement
- Per-command timeout and output caps
- Patch-first edits only (no blind overwrite)

Execution enforcement pipeline (Codex-inspired):
1. Command normalization + segmentation
2. Rules Engine decision (`allow|ask`) on each segment
3. Approval resolution check (if `ask`)
4. Runtime safeguards (hard deny for unsafe/invalid states)
5. Sandbox + network policy + process hardening, then execute

Network policy plane (MVP):
- Local proxy mode for egress governance
- Loopback bind by default
- Allowlist-first, explicit deny overrides allow
- Audit events keep host/method metadata without leaking full sensitive URL payloads

Process hardening baseline (MVP):
- Strip unsafe dynamic loader env vars
- Disable ptrace attach and core dumps where supported
- Enforce minimum process restrictions before tool execution

Brainstorm prompts:
- Should network access be denied by default?
- Is process sandbox enough or container sandbox required?
- What rollback strategy after partial failure?

## 7.5 Provider Router

Responsibilities:
- Choose provider+model+account per turn
- Manage in-provider account switching with two selectable modes
- Track provider/account quota, rate-limit, token usage, and error patterns
- Keep provider interactions stateless from a cache/thread perspective

MVP routing policy:
- Primary provider/model per task category (reasoning, coding, fast-edit)
- Per-provider account switching mode:
  - `turn_robin`: switch account every turn among healthy eligible accounts
  - `quota_failover`: keep sticky account; switch only on rate-limit/quota/availability failures
- Account cooldown on timeout/429/5xx/quota exhausted, then next eligible account
- Max attempts per turn = all eligible accounts in selected provider at turn start
- If all eligible accounts fail, return `provider_accounts_exhausted` and request user action
- User-visible routing decision with provider/model/account and cost hint

Provider cache policy (locked):
- Ignore provider-side cache/session/thread features entirely
- Rebuild prompt every turn from local canonical session state
- Local cache is authoritative (retrieval/context compilation summaries)
- Never assume cache continuity across accounts

Custom provider support (MVP):
- Provider type: `custom-openai-compatible`
- Required config: `name`, `base_url`, `api_key_ref`, `model_map`, `enabled`
- Security constraints: HTTPS-only by default, explicit opt-in for insecure/local endpoints, key redaction in all logs

Multi-account scheduler (MVP):
- Scheduling modes:
  - `turn_robin`
  - `quota_failover`
- Default mode: `quota_failover`
- Eligibility check per turn: enabled + healthy + not cooling down + quota available
- Health/cooldown triggers: timeout, 429, consecutive 5xx
- Auth failure policy: disable account immediately and require re-auth
- Quota-header-missing policy: account remains usable but gets lower priority score
- Guardrails: no account switch mid-stream; switch only at turn boundary
- Model parity rule: all accounts in same provider profile must share identical allowed model set

Mode-specific behavior:
- `turn_robin`:
  - Advance cursor after each successful turn
  - Maintains fairness across healthy accounts
- `quota_failover`:
  - Keep active account sticky while healthy
  - Switch only when active account hits `rate_limited|quota_exhausted|timeout_threshold|auth_invalid`
  - Optional return to primary account after cooldown reset

Routing signals:
- task complexity score
- context size
- tool failure history
- confidence estimate
- provider health (timeout/429/5xx)
- account quota remaining and reset windows
- account token budget and recent token burn
- auth state (valid/expired/revoked)
- quota/rate-limit header availability confidence
- scheduler mode (`turn_robin|quota_failover`)

Brainstorm prompts:
- Should quota_failover switch back eagerly or lazily after reset window?
- What default cooldown windows balance reliability vs throughput?
- Should `provider_accounts_exhausted` auto-pause session or require explicit retry?

## 7.6 Memory System

Responsibilities:
- Preserve short-term task context
- Store reusable long-term project preferences
- Act as canonical conversation state for provider/account switching continuity

MVP layers:
- Session memory: rolling summaries + key facts
- Project memory: conventions, architecture notes, preferred commands

Controls:
- User can inspect/edit/delete memories
- Memory entries carry provenance (source event)

Brainstorm prompts:
- What retention policy balances usefulness vs drift?
- How to protect against bad memory writes?
- Should memory be branch-aware?

## 7.7 Policy Engine

Responsibilities:
- Assign risk score for each proposed action
- Decide auto-allow or require-approval based on runtime mode
- Enforce workspace runtime profile and mode
- Stay ask-first for user-governed behavior; do not enforce hard technical rejects

Risk tiers:
- Low: read/search, safe test command
- Medium: file modifications, package install
- High: destructive commands, credential or network-sensitive operations

Policy operation modes (MVP):
- `ask` (default): dangerous/high-risk actions go to approval flow
- `auto_allow_all`: no approval prompts; low/medium/high actions are policy-allowed

Mode behavior contract:
- `low` risk => `allow`
- `medium` risk =>
  - `ask` mode: `ask`
  - `auto_allow_all` mode: `allow`
- `high` risk =>
  - `ask` mode: `ask`
  - `auto_allow_all` mode: `allow`
- Runtime safeguards still execute and may deny unsafe/invalid operations regardless of mode

Brainstorm prompts:
- Should `auto_allow_all` be session-only or workspace-persistent by default?
- Should risky actions in `auto_allow_all` require explicit mode banner in UI?
- What runtime safeguards are non-overridable regardless of mode?

Approval lifecycle (Codex-inspired):
- States: `pending -> resolved -> completed`
- Resolution decisions (MVP): `accept_once`, `accept_for_session`, `decline`, `cancel`
- Optional amendments (MVP):
  - `exec_rule_amendment` (narrow prefix rule)
  - `network_rule_amendment` (host-level allow/deny)
- Every resolution emits explicit event and is linked to the triggering request

### Runtime Safeguards Layer (Hard Deny)

Purpose:
- Enforce non-overridable execution safety for unsafe or invalid operations.
- Separate technical safety rejects from user policy prompts.

Decision model:
- `allow | deny` (this is the only deny path in v1)

Safeguard codes (MVP):
- `OUTSIDE_WORKSPACE_PATH`
- `COMMAND_SEGMENT_PARSE_FAILED`
- `APPROVAL_TOKEN_MISSING`
- `EXECUTABLE_PATH_SPOOF`
- `MID_STREAM_ACCOUNT_SWITCH`
- `PATCH_STRUCTURAL_INVALID`
- `PROVIDER_ENDPOINT_UNTRUSTED`
- `NETWORK_POLICY_DENIED`
- `TOOL_CONTRACT_INVALID`
- `PROCESS_HARDENING_INIT_FAILED`
- `SANDBOX_ENFORCEMENT_FAILED`
- `UNSAFE_ENV_INJECTION`

Audit requirements:
- Every reject returns `{safeguard_code, reason, remediation_hint}`
- Rejections emit event `runtime_safeguard.rejected`

### 7.7.1 Policy Defaults: `solo-default-v1` (MVP)

Policy defaults are ask-first with explicit allow rules for low-risk read/test operations.
When `dangerous_action_mode=auto_allow_all`, all risk tiers are policy-allowed (no approval prompts). Runtime safeguards remain non-overridable.

```yaml
ruleset:
  id: solo-default-v1
  mode: ask-first
  dangerous_action_mode: ask # switchable to auto_allow_all
  precedence:
    - base_rules_strictest: ask_over_allow
    - session_grant_override: enabled

rules:
  # Read/Search safe
  - id: R001
    match: { tool: read }
    decision: allow
    reason: "Read-only safe operation"

  - id: R002
    match: { tool: grep }
    decision: allow
    reason: "Search-only safe operation"

  - id: R003
    match: { tool: glob }
    decision: allow
    reason: "Path discovery safe operation"

  # File mutations
  - id: R010
    match: { tool: edit_patch }
    decision: ask
    reason: "Code mutation requires user approval"

  - id: R011
    match: { tool: edit_write }
    decision: ask
    reason: "Direct overwrite must be reviewed"

  # Git read-only
  - id: R020
    match: { tool: git, command_prefix: "git status" }
    decision: allow
    reason: "Read-only git command"

  - id: R021
    match: { tool: git, command_prefix: "git diff" }
    decision: allow
    reason: "Read-only git command"

  - id: R022
    match: { tool: git, command_prefix: "git log" }
    decision: allow
    reason: "Read-only git command"

  # Git mutating / remote side-effects
  - id: R030
    match: { tool: git, command_prefix: "git add" }
    decision: ask
    reason: "Staging modifies index"

  - id: R031
    match: { tool: git, command_prefix: "git commit" }
    decision: ask
    reason: "Creates repository history"

  - id: R032
    match: { tool: git, command_prefix: "git push" }
    decision: ask
    reason: "Remote side-effect requires explicit consent"

  # Shell safe tests/lint
  - id: R040
    match: { tool: shell, command_prefix: "npm test" }
    decision: allow
    reason: "Test command"

  - id: R041
    match: { tool: shell, command_prefix: "pnpm test" }
    decision: allow
    reason: "Test command"

  - id: R042
    match: { tool: shell, command_prefix: "pytest" }
    decision: allow
    reason: "Test command"

  - id: R043
    match: { tool: shell, command_prefix: "npm run lint" }
    decision: allow
    reason: "Lint command"

  # Build/install/network-like shell
  - id: R050
    match: { tool: shell, command_prefix: "npm run build" }
    decision: ask
    reason: "Potential heavy side-effect"

  - id: R051
    match: { tool: shell, command_prefix: "pnpm build" }
    decision: ask
    reason: "Potential heavy side-effect"

  - id: R052
    match: { tool: shell, command_prefix: "npm install" }
    decision: ask
    reason: "Dependency graph mutation"

  - id: R053
    match: { tool: shell, command_prefix: "pnpm add" }
    decision: ask
    reason: "Dependency graph mutation"

  - id: R054
    match: { tool: shell, command_prefix: "curl " }
    decision: ask
    reason: "Network egress action"

  # Catch-all
  - id: R999
    match: { tool: "*" }
    decision: ask
    reason: "Unknown operation defaults to ask"
```

Validation vectors (MVP baseline):
- Match:
  - `git status` => `allow` (R020)
  - `git diff --staged` => `allow` (R021)
  - `git add .` => `ask` (R030)
  - `git commit -m "x"` => `ask` (R031)
  - `pnpm test` => `allow` (R041)
  - `npm install` => `ask` (R052)
- Non-match / safeguard:
  - `git status && rm -rf /` => segment 2 must not inherit allow
  - unknown shell command => fallback `ask` (R999)
  - symlink escape outside workspace => `OUTSIDE_WORKSPACE_PATH`
  - missing approval token for ask-gated action => `APPROVAL_TOKEN_MISSING`

`auto_allow_all` mode behavior:
- When risk tier is `low`, `medium`, or `high`, action is policy-allowed with no approval prompt
- Runtime safeguards still may reject with `runtime_safeguard.rejected`
- UI must show explicit mode badge to prevent surprise side effects

## 7.8 Storage Layer

Responsibilities:
- Persist sessions, events, indexes, patches, approvals
- Support fast local queries and reliable recovery

MVP stack:
- SQLite (WAL mode)
- SQLite vector extension or sidecar vector store
- Artifact files for large logs/diffs

Reliability:
- Append-only event log
- Snapshot checkpoints every N events
- Startup recovery via event replay

Brainstorm prompts:
- Encrypt DB at rest in MVP or phase 2?
- How to compact old events without losing auditability?
- Best backup/restore UX for local dev?

## 7.9 Plugin and MCP Integration

Responsibilities:
- Extend agent with external tools/resources
- Keep extension model controlled and auditable

MVP approach:
- MCP client support for trusted local/known servers
- Permission prompt on first use per server capability

Brainstorm prompts:
- Plugin signing needed before public beta?
- Per-plugin resource quotas?
- Sandboxing boundary for untrusted plugins?

## 7.10 Evaluation and Telemetry

Responsibilities:
- Measure real task success and failure modes
- Provide regression signals for releases

MVP metrics:
- Task success rate
- Tool error rate by category
- Patch acceptance/rejection rate
- Median and P95 end-to-end latency
- Recovery success after daemon restart

Telemetry policy:
- Local-only logs by default
- Optional anonymized telemetry opt-in

Brainstorm prompts:
- What benchmark mix reflects actual user tasks?
- How many failed runs trigger release block?
- Which metrics should be visible in product UI?

## 8) Data Model Blueprint

Core entities:
- `workspaces(id, path, created_at, policy_profile)`
- `sessions(id, workspace_id, title, status, created_at, updated_at)`
- `messages(id, session_id, role, content, token_count, created_at)`
- `plans(id, session_id, status, confidence, created_at)`
- `plan_steps(id, plan_id, step_no, description, status, started_at, ended_at)`
- `tool_calls(id, session_id, step_id, tool_name, input_json, output_ref, status, risk, created_at)`
- `approvals(id, approval_request_id, tool_call_id, decision, actor, rationale, created_at)`
- `approval_requests(id, session_id, request_type, status, scope, payload_json, created_at, resolved_at)`
- `file_patches(id, session_id, file_path, diff_text_ref, apply_status, created_at)`
- `index_chunks(id, workspace_id, file_path, symbol, hash, updated_at)`
- `embeddings(id, chunk_id, model_id, vector_ref, created_at)`
- `memories(id, workspace_id, scope, key, value, provenance_event_id, updated_at)`
- `model_runs(id, session_id, phase, provider_id, account_id, model_id, latency_ms, tokens_in, tokens_out, quality_score, created_at)`
- `events(id, session_id, type, payload_json, created_at)`
- `providers(id, name, auth_type, enabled, config_json, created_at, updated_at)`
- `provider_credentials(id, provider_id, credential_ref, scopes_json, created_at, updated_at)`
- `provider_health(id, provider_id, window_start, success_rate, p95_latency_ms, rate_limit_hits, error_count, updated_at)`
- `provider_routing_config(provider_id, switch_mode, sticky_account_id, updated_at)`
- `provider_accounts(id, provider_id, account_name, auth_type, credential_ref, enabled, priority, created_at, updated_at)`
- `provider_account_health(id, account_id, window_start, success_rate, p95_latency_ms, timeout_count, rate_limit_hits, auth_errors, cooldown_until, updated_at)`
- `provider_quota_snapshots(id, account_id, source_ts, quota_limit, quota_remaining, reset_at, created_at)`
- `provider_rate_limit_snapshots(id, account_id, source_ts, requests_limit, requests_remaining, requests_reset_at, tokens_limit, tokens_remaining, tokens_reset_at, created_at)`
- `provider_token_usage(id, account_id, session_id, turn_id, tokens_in, tokens_out, cost_estimate, created_at)`
- `turn_account_assignments(id, session_id, turn_id, provider_id, account_id, reason, created_at)`
- `agents(id, name, type, mode, enabled, source, manifest_json, created_at, updated_at)`
- `agent_permissions(id, agent_id, permission_json, updated_at)`
- `skills(id, name, version, source, enabled, manifest_json, created_at, updated_at)`
- `agent_profiles(id, workspace_id, name, config_json, created_at, updated_at)`
- `rulesets(id, workspace_id, scope, version, created_at, updated_at)`
- `rules(id, ruleset_id, priority, match_json, decision, constraints_json, reason, created_at)`
- `rule_test_vectors(id, ruleset_id, vector_type, input_text, expected_decision, created_at)`
- `workflow_templates(id, name, version, definition_json, created_at, updated_at)`
- `workflow_runs(id, session_id, workflow_template_id, status, started_at, ended_at)`
- `workflow_run_steps(id, workflow_run_id, step_id, status, started_at, ended_at, output_ref)`
- `slash_commands(id, command, workflow_template_id, runtime_mode, enabled, updated_at)`
- `commands(id, name, source, template_ref, enabled, updated_at)`
- `tools(id, name, source, risk_level, schema_json, enabled, updated_at)`
- `plugins(id, name, source, version, enabled, manifest_json, updated_at)`
- `hooks(id, plugin_id, hook_name, enabled, timeout_ms, config_json, created_at, updated_at)`
- `hook_events(id, session_id, hook_id, status, latency_ms, payload_ref, created_at)`
- `runtime_safeguard_rejections(id, session_id, tool_call_id, safeguard_code, reason, remediation_hint, created_at)`
- `agent_team_runs(id, session_id, status, created_at, completed_at)`
- `agent_team_members(id, team_run_id, agent_id, parent_agent_id, status, started_at, completed_at)`
- `agent_team_messages(id, team_run_id, from_agent_id, to_agent_id, message_ref, created_at)`
- `workflow_step_bindings(id, workflow_template_id, step_id, role_agent_id, skill_id, tool_scope_json, created_at, updated_at)`
- `registry_versions(id, registry_name, version, changed_by, changed_at)`

Data design notes:
- Every mutable action emits an event.
- Large payloads (stdout/diff) stored as artifact files with references.
- Branch name/version can be added to workspace-scoped tables for branch-aware behavior.

## 9) API and IPC Blueprint

Transport options:
- Unix socket / Named pipe (preferred for local trusted clients)
- stdio JSONL profile (compatibility mode for embedded/hosted clients)
- localhost HTTP with short-lived auth token fallback

Protocol lifecycle:
- Clients must call `runtime.initialize` then `runtime.initialized` before mutating calls
- Capability negotiation happens during initialize handshake
- Overload/backpressure responses return structured retry hints (e.g., retry_after_ms)

Core RPC methods:
- `runtime.initialize(client_info, capabilities)`
- `runtime.initialized()`
- `session.create(workspace_path)`
- `session.resume(session_id)`
- `message.send(session_id, text)`
- `plan.get(session_id)`
- `approval.pending(session_id)`
- `approval.resolve(request_id, decision, scope, amendment, note)`
- `patch.list(session_id)`
- `patch.apply(patch_id)`
- `tool.replay(tool_call_id)`
- `workspace.reindex(workspace_id, mode)`
- `diagnostics.health()`
- `registry.snapshot(registry_name)`
- `registry.validate(registry_name)`
- `registry.validate.role_skill_boundary(agent_id | all)`
- `registry.validate.capability(workflow_template_id)`
- `registry.rollback(registry_name, version)`
- `runtime.profile.get()`
- `runtime.mode.get(session_id)`
- `runtime.mode.set(session_id, mode)`
- `agent.list()`
- `agent.upsert(agent_manifest)`
- `agent.enable(agent_id)` / `agent.disable(agent_id)`
- `skill.list()`
- `skill.upsert(skill_manifest)`
- `skill.enable(skill_id)` / `skill.disable(skill_id)`
- `workflow.list()`
- `workflow.upsert(template)`
- `workflow.bind_step(workflow_template_id, step_id, role_agent_id, skill_id, tool_scope)`
- `workflow.validate_bindings(workflow_template_id)`
- `workflow.run(session_id, workflow_template_id, args)`
- `command.list()`
- `command.upsert(command_spec)`
- `command.delete(command_id)`
- `slash.execute(session_id, command_line)`
- `tool.list()`
- `tool.register(tool_spec)` / `tool.unregister(tool_name)`
- `plugin.list()`
- `plugin.install(source)`
- `plugin.enable(plugin_id)` / `plugin.disable(plugin_id)`
- `policy.explain(action_preview)`
- `policy.validate(ruleset_id)`
- `policy.simulate(action_preview)`
- `hooks.list()`
- `hooks.toggle(hook_id, enabled)`
- `hooks.test(hook_id, sample_payload)`
- `team.spawn(session_id, agent_id, task)`
- `team.send(team_member_id, input)`
- `team.resume(team_member_id)`
- `team.wait(team_run_id | team_member_id)`
- `team.close(team_member_id)`
- `provider.list()`
- `provider.connect(provider_name, auth_mode)`
- `provider.disconnect(provider_id)`
- `provider.health()`
- `provider.add_custom(config)`
- `provider.update_custom(provider_id, config)`
- `provider.account.list(provider_id)`
- `provider.account.add(provider_id, account_config)`
- `provider.account.disable(account_id)` / `provider.account.enable(account_id)`
- `provider.switch_mode.get(provider_id)` / `provider.switch_mode.set(provider_id, mode)`
- `provider.usage(account_id, window)`

Event stream (server push):
- `runtime.initialized`
- `orchestrator.state_changed`
- `plan.updated`
- `tool.started`
- `tool.completed`
- `approval.required`
- `approval.resolved`
- `patch.generated`
- `verification.completed`
- `session.completed`
- `workflow.started`
- `workflow.step_changed`
- `agent.updated`
- `skill.loaded`
- `command.updated`
- `tool.updated`
- `plugin.updated`
- `registry.validated`
- `registry.rolled_back`
- `slash.executed`
- `policy.decision_explained`
- `hook.called`
- `hook.failed`
- `runtime_safeguard.rejected`
- `team.member_spawned`
- `team.member_completed`
- `team.message`
- `provider.account_switched`
- `provider.quota_updated`
- `provider.rate_limit_updated`
- `provider.token_usage_updated`
- `runtime.overloaded`

## 10) End-to-End Sequence (Primary Flow)

```text
User -> Adapter: request task
Adapter -> Daemon: runtime.initialize + runtime.initialized
Adapter -> Daemon: message.send
Daemon/Orchestrator -> Context Engine: retrieve relevant context
Orchestrator -> Provider Router: choose provider/model/account (`turn_robin` or `quota_failover`) and generate next action
Orchestrator -> Policy Engine: risk evaluation
Policy Engine -> Orchestrator: allow / require approval
Orchestrator -> Adapter: approval.required (when needed)
Adapter -> Orchestrator: approval.resolve
Orchestrator -> Runtime Safeguards: final execution safety check (non-overridable, when action proceeds)
Orchestrator -> Tool Runtime: execute approved action (when action proceeds)
Tool Runtime -> Orchestrator: structured output
Orchestrator -> Verification: run tests/lint (if policy enabled)
Orchestrator -> Adapter: patch + summary + confidence
Adapter -> User: review and apply
```

### 10.1 End-to-End Sequence (Agent Team Flow)

```text
User -> Adapter: run team task
Adapter -> Daemon: team.spawn(primary_agent, task)
Orchestrator -> Team Orchestrator: spawn sub-agents by workflow step
Team Orchestrator -> Agent Runtime: assign WHO (agent role)
Agent Runtime -> Skills Registry: load HOW (skill procedure)
Sub-agents -> Tool Runtime: execute scoped tasks
Team Orchestrator -> Adapter: stream member status/messages
Team Orchestrator -> Orchestrator: aggregate outputs
Orchestrator -> Adapter: consolidated result + patch plan
```

## 11) Security, Privacy, and Trust Blueprint

Threat model highlights:
- Prompt injection from repository files
- Destructive shell command execution
- Secret leakage in prompts/logs
- Plugin abuse or privilege escalation

Security controls:
- Prompt sanitization + source tagging
- Secret redaction before model input and logs
- Policy gate on all side-effectful actions
- Strict cwd confinement for tool runtime
- Optional network egress restriction
- Approval requirement for high-risk actions in `ask` mode
- Provider endpoint validation (HTTPS default, blocked private-address targets unless explicitly trusted)
- Credentials in OS keychain/secure store only; DB stores references
- Local network policy proxy uses loopback bind by default
- Explicit deny beats allow inside network policy evaluation
- Process hardening strips unsafe loader env vars and disables risky debugging hooks where supported
- Prefer additional narrow permissions/amendments before broad escalation

Privacy controls:
- Local control-plane by default (session/tooling/index/audit)
- Explicit provider consent on first outbound usage (BYOK/OAuth)
- Outbound context minimization + secret redaction + provider boundary rules
- Provider cache/thread metadata is not used or persisted
- Provider telemetry ingest limited to quota, rate-limit, and token usage
- Separate toggles: telemetry, provider inference, plugin network

## 12) Reliability and Failure Handling

Failure classes and handling:
- Provider timeout/rate-limit: retry with bounded backoff across eligible accounts in same provider
- In `quota_failover` mode: switch active account only on quota/rate-limit/failure triggers
- In `turn_robin` mode: rotate cursor each turn and skip unhealthy accounts
- Eligible accounts exhausted: return `provider_accounts_exhausted`, park turn, request user action
- OAuth token expiry/revocation: refresh flow or re-auth prompt with safe pause
- Tool failure: bounded retries, then escalate with diagnostics
- Index corruption: rebuild incremental, fallback lexical only
- Daemon crash: replay events and restore last stable state
- Approval deadlock: notify user and park session safely

SLO targets:
- Daemon startup < 2 s
- Session resume < 1.5 s
- Tool execution timeout default 120 s
- Crash recovery complete < 5 s for normal sessions

## 13) Deployment and Packaging Blueprint

Modes:
- Single-user local workstation (default)
- Team-managed install is post-MVP
- Enterprise self-host is post-MVP

Packaging:
- Go binaries: `agentd` (daemon) + `agent` (CLI/TUI)
- Local web UI assets embedded in daemon binary (`go:embed`) for single-process local serving
- Auto-update channel with rollback support
- Version compatibility matrix between client adapters and daemon
- Cross-platform release automation with signed artifacts and checksums

## 14) Performance Strategy

- Incremental indexing only; avoid full reindex unless required
- Cache retrieval results per step with invalidation rules
- Keep prompt compilation/cache local; do not rely on provider-side cache hits
- Stream partial model outputs and tool logs
- Preload frequently used metadata on daemon startup
- Apply adaptive context compression for large repositories

## 15) Testing and Evaluation Strategy

Test layers:
- Unit: planner logic, risk scorer, index ranking functions
- Integration: orchestrator + tools + policy gates
- E2E: real repository task suites with acceptance checks
- Chaos tests: kill daemon during execution, replay validation

Release gates:
- No critical security regression
- Benchmark success above threshold
- Latency and reliability within SLO budgets

## 16) Build Roadmap (12 Weeks MVP)

Phase 0 (Week 0): discovery and scope lock
- finalize user persona, acceptance benchmarks, policy baseline

Phase 1 (Week 1-2): daemon skeleton + CLI
- session lifecycle, event bus, health checks

Phase 2 (Week 3-4): indexing + retrieval
- file watcher, chunking, hybrid search

Phase 3 (Week 5-6): orchestrator + policy + tool runtime
- step loop, approvals, safe command execution

Phase 4 (Week 7-8): control layer + patch workflow
- agent/skill/workflow/command/tool/plugin/hook registries v1
- patch generation, diff review contracts, verification hooks

Phase 5 (Week 9-10): local web UI adapter
- session timeline, approval dock, team execution view, patch apply flow

Phase 6 (Week 11-12): provider resiliency hardening
- scheduler hardening (`turn_robin` + `quota_failover`), quota/rate-limit telemetry, auth lifecycle robustness, provider-exhausted UX, eval suite stabilization, packaging

## 17) Build vs Buy Decisions

Candidate buy/adopt:
- CLI framework: `cobra`
- TUI framework: `bubbletea` + `bubbles` + `lipgloss`
- API transport: Connect RPC (`connect-go`) + protobuf contracts
- SQLite access tooling: `sqlc` + migration tool
- File watcher: `fsnotify`
- OAuth: `golang.org/x/oauth2` (PKCE/device-code)
- MCP protocol client implementation
- Parsing/symbol extraction: tree-sitter + LSP where available

Build in-house:
- Orchestrator policies and state machine
- Tool safety model and approval UX contracts
- Session/event schema and recovery model
- Task benchmark harness aligned with product goals
- Provider router and multi-account schedulers (`turn_robin`, `quota_failover`)
- Role-skill boundary validators and workflow binding validators

### 17.1 Minimal Go Project Layout

```text
cmd/
  agentd/           # daemon entry
  agent/            # CLI/TUI entry
internal/
  api/              # RPC handlers + contracts bridge
  orchestrator/
  teams/
  context/
  runtime/
  policy/
  safeguards/
  provider/
  registry/
  storage/
  web/
pkg/
  contract/         # shared protobuf/domain DTOs
db/
  migrations/
web/
  dist/             # built local web UI assets (embedded)
```

## 18) Workshop Template for Brainstorming Each Module

Use this template per module session (45-60 min):
- Problem statement for module
- Option A/B/C with explicit trade-offs
- Proposed default decision
- Risks and mitigations
- Open questions (max 5)
- Decision log (ADR id, owner, due date)

Suggested module order:
1. Tool Runtime + Policy Engine
2. Context Engine
3. Orchestrator
4. Provider Router
5. Agents + Skills registries
6. Commands + Tools + Plugins + Hooks
7. Agent Team Orchestrator
8. Memory System
9. Adapter UX (TUI + local web UI)
10. Eval and Telemetry

## 19) Priority Backlog (Initial)

P0:
- Go module skeleton (`cmd/agentd`, `cmd/agent`, `internal/*`) and shared contracts
- Daemon skeleton, IPC, event log
- Safe tool runtime with risk tiering
- Hybrid retrieval minimal viable quality
- Patch preview/apply cycle
- Control layer skeleton (agents, skills, workflows, slash commands, tools, plugins, hooks)
- Custom provider adapter (`base_url + api_key`) with endpoint validation
- Provider account pool + dual switching modes (`turn_robin`, `quota_failover`) + quota/rate-limit/token telemetry
- Agent team orchestration primitives (`spawn/send/resume/wait/close`)

P1:
- TUI UX polish + local web UI polish
- Session memory controls
- Benchmark harness and CI reliability checks
- Skills/Agents/Commands customization UX
- Team execution timeline in local web UI

P2:
- Plugin/MCP ecosystem hardening
- Advanced routing and adaptive retries
- Optional session share and advanced team controls
- Hook sandbox hardening and signed plugin policy
- Cross-provider execution failover (post-v1)

## 20) Resolved Defaults (Build Freeze Inputs)

1. Provider onboarding defaults to BYOK first; OAuth optional per provider.
2. `provider_accounts_exhausted` pauses turn and prompts user action.
3. Local web UI is first-class and served by daemon.
4. Default runtime mode is `ask`; optional `auto_allow_all` is session-scoped unless user persists it.
5. Commit allowed with `ask`; push allowed with `ask` and explicit confirmation.
6. Slash commands inherit active runtime mode.
7. Skill activation trust defaults to per-session approval unless explicitly persisted.
8. Hook fail-closed set: `before_model_call`, `before_tool_call`, `before_patch_apply`, `permission_ask`.
9. Multi-account switching modes are both supported in v1: `turn_robin`, `quota_failover`.
10. Agent contract is locked: Agent=WHO, Skill=HOW.
11. Default provider switch mode is `quota_failover`; users can opt into `turn_robin`.
12. Workflow steps must bind both WHO (agent role) and HOW (skill), else invalid.
13. Agents cannot contain procedural step logic; skills cannot carry ownership authority.
14. Registry changes require validation before activation and support version rollback.
15. API calls require `runtime.initialize` handshake before mutating operations.
16. Runtime implementation language is Go; clients communicate over typed local RPC.

## 21) Decision Log Seed (ADR Starters)

- ADR-001: choose adapter + daemon architecture
- ADR-002: tool runtime safety and command policy
- ADR-003: retrieval architecture and indexing granularity
- ADR-004: provider routing and in-provider account-rotation policy (no cross-provider in v1)
- ADR-005: memory retention and user controls
- ADR-006: telemetry and privacy defaults
- ADR-007: control layer contracts (agents/skills/workflows/slash/tools/plugins/hooks)
- ADR-008: provider cache independence + dual switching modes (`turn_robin`, `quota_failover`)
- ADR-009: multi-agent team orchestration lifecycle and limits
- ADR-010: native runtime separation contract (`Agent=WHO`, `Skill=HOW`) and workflow binding rules
- ADR-011: Go runtime stack and RPC handshake contract

## 22) Reference Alignment (Codex + OpenCode)

Codex-aligned baseline kept near-identical:
- Server-first local daemon + typed protocol boundary + thread/turn lifecycle
- Layered runtime safety: exec rules, sandbox, network policy, process hardening
- Approval lifecycle semantics and auditable event model
- Multi-agent orchestration primitives and guardrail limits

OpenCode-aligned baseline kept near-identical:
- Registry-driven customization for agents/skills/commands/tools/plugins/hooks
- Local web workflow with daemon-served UI and realtime event streaming
- Strong extension ergonomics (workspace/user/global overrides)
- Tool + plugin + hook composability for custom product behavior

Intentional project-specific decisions:
- Provider-only inference with BYOK/OAuth
- Multi-account switching modes: `turn_robin` and `quota_failover`
- Agent contract locked as WHO, Skill contract locked as HOW
- One simple runtime profile with explicit mode toggle (`ask`, `auto_allow_all`)
- Stronger default hardening than OpenCode baseline (runtime safeguards + process/network controls)
- No external kit import pipeline in MVP (native registries are source of truth)

## 23) Native Runtime Contract (No Import Pipeline)

Design intent:
- Runtime definitions are authored and managed natively through registries and APIs.
- External kits can inspire content, but runtime behavior is not coupled to import pipelines.

Core contract:
- `Agent = WHO`
  - Identity, ownership, authority, delegation limits, escalation rights
  - Never stores step-by-step procedures
- `Skill = HOW`
  - Procedural method, reusable instructions, tool recipes, validation checkpoints
  - Never stores ownership/authority semantics

Workflow execution contract:
- Every workflow step must bind:
  - `role_agent_id` (WHO executes)
  - `skill_id` (HOW to execute)
  - `tool_scope` (allowed tools for this step)
- A step missing any binding is invalid and cannot run.

Runtime validators (required):
- `role_skill_boundary_lint`
  - Reject if agent manifest contains procedural checklist blocks
  - Reject if skill content defines ownership/escalation authority
- `workflow_binding_lint`
  - Reject if workflow step has missing or conflicting role/skill bindings
- `capability_lint`
  - Reject if workflow/tool scope exceeds role authority envelope

Recommended execution sequence:
1. Select workflow template
2. Bind step to WHO + HOW (+ tool scope)
3. Resolve permissions and runtime mode (`ask` or `auto_allow_all`)
4. Execute via tool runtime and safeguards
5. Verify outputs and aggregate team results

Native customization surfaces (all first-class):
- Agents Registry
- Skills Registry
- Workflow Registry
- Slash Command Registry
- Tool Registry
- Plugin Runtime
- Hook Bus
- Team Orchestrator

Minimum acceptance tests:
- create/update custom agent with authority envelope and validate
- create/update custom skill and validate procedural scope
- run workflow with explicit WHO/HOW binding per step
- reject workflow step missing role or skill
- reject skill that attempts to override runtime safeguards
- verify multi-agent team run honors depth/thread/runtime limits

---

This blueprint is intentionally practical: prioritize trustworthy execution and reviewable changes first, then scale autonomy and ecosystem depth.
