# CLAUDE.md — Local Agent SOP Entry Point

This workspace uses a file-based, local-first pipeline with shared context, structured handoff contracts, and phased execution.

## Pipeline

`/kd-brainstorm -> /kd-handoff-spec -> /kd-dev -> /kd-qa -> /kd-handoff-dev -> /kd-release -> /kd-content`

## Source of Truth

- Orchestrator rules: `AGENTS.md`
- Operational guide: `README.md`
- Shared context: `_context/`
- Work queue: `_handoff/queue/`
- Handoff contract rules: `_handoff/README.md`
- Lessons learned: `_context/lessons.md`

---

## Agent Behavior Rules

These rules apply to ALL agents at ALL times, regardless of pipeline stage.

### 1. Plan Before You Build
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions).
- If something goes sideways, **STOP and re-plan immediately** — never keep pushing a failing approach.
- Use plan mode for verification steps, not just implementation.
- Write detailed specs upfront to reduce ambiguity.

### 2. Subagent Strategy
- Use subagents liberally to keep the main context window clean.
- Offload research, exploration, and parallel analysis to subagents.
- For complex problems, throw more compute at it via subagents.
- One task per subagent for focused execution.

### 3. Self-Improvement Loop
- After ANY correction from the user: update `_context/lessons.md` with the pattern.
- Write rules for yourself that prevent the same mistake from recurring.
- Ruthlessly iterate on these lessons until the mistake rate drops.
- Review `_context/lessons.md` at session start for the relevant project.

### 4. Verification Before Done
- Never mark a task complete without proving it works.
- Diff behavior between main and your changes when relevant.
- Ask yourself: *"Would a staff engineer approve this?"*
- Run tests, check logs, demonstrate correctness.

### 5. Demand Elegance (Balanced)
- For non-trivial changes: pause and ask *"Is there a more elegant way?"*
- If a fix feels hacky: *"Knowing everything I know now, implement the elegant solution."*
- Skip this for simple, obvious fixes — do not over-engineer.
- Challenge your own work before presenting it.

### 6. Autonomous Bug Fixing
- When given a bug report: just fix it. Do not ask for hand-holding.
- Point at logs, errors, failing tests — then resolve them.
- Zero context switching required from the user.
- Go fix failing CI tests without being told how.

### 7. Simplicity & Minimal Impact
- Make every change as simple as possible. Touch minimal code.
- Find root causes. No temporary fixes. Senior developer standards.
- Changes should only touch what is necessary. Avoid introducing bugs.

---

## Pipeline Modes

### Feature Pipeline (default)
Full pipeline for new features, improvements, and non-trivial changes:
```
/kd-brainstorm → /kd-handoff-spec → /kd-dev → /kd-qa → /kd-handoff-dev → /kd-release → /kd-content
```

### Bug Fix Fast Path
For bug reports and failing tests — skip brainstorm and spec, go straight to fix:
```
Bug report → /kd-dev → /kd-qa → /kd-handoff-dev → /kd-release [→ /kd-content if user-facing]
```
Rules for fast path:
- Only for clear bugs with reproducible symptoms (failing tests, error logs, user-reported defects).
- Dev agent creates a minimal handoff ticket directly (no spec required).
- Bug fixes use `spec: none`. All downstream stages handle this explicitly — no spec required.
- Must still pass full QA gates — no shortcuts on verification.
- If the bug reveals a deeper architectural issue, escalate to the full Feature Pipeline.

### Refactor / Tech Debt Path
For internal quality improvements with no user-facing changes:
```
/kd-brainstorm → /kd-handoff-spec → /kd-dev → /kd-qa → /kd-handoff-dev → /kd-release
```
Rules:
- Spec must explicitly state "no user-facing change" and define regression criteria.
- QA must verify behavior preservation (before/after comparison).
- No content handoff — internal changes don't need marketing.

### Dependency / Security Upgrade Path
For package bumps, CVE patches, framework upgrades:
```
/kd-dev → /kd-qa → /kd-handoff-dev → /kd-release
```
Rules:
- Dev creates a minimal ticket with: package name, old→new version, changelog summary, compatibility notes.
- QA focuses on regression testing and build verification.
- Must include rollback instructions (pin to old version).

---

## Core Principles

1. **Fact-first** — Gather and verify facts before brainstorming solutions
2. **Phased execution** — Break work into independently testable phases (required for effort ≥ M)
3. **Contract-driven handoffs** — Every handoff includes task_description, acceptance_criteria, context_keys, and output_mode
4. **Guardrailed specs** — Automated validation before specs enter the dev pipeline
5. **Loop detection** — Progress Ledger prevents infinite QA→dev cycles (escalate at loop_count ≥ 3)
6. **Self-improvement** — Every correction becomes a lesson; every lesson prevents future mistakes

---

## Stage Rules

Each stage has detailed workflow instructions in its skill file (`.agents/skills/kd-*/SKILL.md`). The CLAUDE.md and AGENTS.md provide the behavioral framework; skill files provide stage-specific procedures.

**Key stage contracts:**
- **brainstorm** → produces draft spec + research note
- **handoff-spec** → validates spec, creates handoff ticket with Contract section
- **dev** → implements current phase only, self-verifies
- **qa** → runs checks, fills Progress Ledger, routes PASS/FAIL with `failure_class`
- **handoff-dev** → routes next phase or creates release ticket
- **release** → presents deploy command (never auto-deploys), creates content ticket
- **content** → generates artifacts, archives everything
---

## File Standards

### Spec frontmatter

```yaml
---
id: SPEC-XXX
title: Feature Title
status: draft|approved|implemented|released|archived
priority: P0|P1|P2
effort: S|M|L
created: YYYY-MM-DD
author: agent:brainstorm|dev|qa|release|content
---
```

### Handoff frontmatter

```yaml
---
id: HO-XXX
from: brainstorm|dev|qa|release
to: dev|qa|release|content
priority: P0|P1|P2
status: pending|in-progress|done|blocked|needs-respec|cancelled|release-failed
created: YYYY-MM-DDTHH:MM:SSZ
spec: SPEC-XXX
total_phases: N
current_phase: N
loop_count: 0
failure_class: none|code|spec|env|flake
output_mode: full_history|last_message
---
```

## Quality Gates

- Backend: `go vet ./...`, `go test ./...`, `go build ./...`
- Frontend: `npm run build`, `npm test`, `npm run test:e2e`
