# Design Guidelines

## UX Principles
- Keep control explicit: user sees what will run, why, and with what risk
- Keep context visible: show active workflow step, role (WHO), skill (HOW)
- Keep interruptions minimal but safe: ask when needed, otherwise continue

## CLI/TUI Guidelines
- Single focused task pane + compact event timeline
- Approval prompts must include:
  - action summary
  - risk level
  - policy reason
  - scope (`once` or `session`)
- Show runtime mode badge (`ask` / `auto_allow_all`) at all times

## Local Web UI Guidelines
- First-class local dashboard served by daemon
- Required panels:
  - session timeline
  - approvals queue
  - patch preview/apply
  - provider/account status
  - team execution view

## Multi-Agent Team UX
- Each team member row shows:
  - role/agent
  - current skill
  - state (running/waiting/failed/completed)
- Display message flow between members in chronological order

## Safety UX
- Never hide runtime safeguard rejections
- Rejections must include `safeguard_code` and remediation hint
- If in `auto_allow_all`, show persistent warning banner
