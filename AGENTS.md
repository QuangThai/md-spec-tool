# AGENTS.md — md-spec-tool Workspace Orchestrator

## Pipeline Overview

`/kd-brainstorm -> /kd-handoff-spec -> /kd-dev -> /kd-qa -> /kd-handoff-dev -> /kd-release -> /kd-content`


## Workspace Structure
```
Workspace/
├── _context/               # Shared context layer (symlinked across repos)
│   ├── product-state.md    # Current product state & priorities
│   ├── specs/              # Feature specs (draft → approved → implemented → released)
│   ├── decisions/          # Architecture & product decision records
│   ├── research/           # Research notes & competitive analysis
│   ├── design/             # Design docs & UX decisions
│   ├── metrics/            # Quality metrics & KPIs
│   ├── content/            # Generated content (blog, changelog, social)
│   └── lessons.md          # Self-improvement log (cross-session learning)
├── _handoff/               # Inter-agent work queue
│   ├── queue/              # Pending handoff tickets
│   └── archive/            # Completed handoffs
├── .agents/skills/         # Pipeline skills (kd-*)
├── docs/                   # Deploy scripts & operational docs
├── backend/                # Go/Gin backend
└── frontend/               # Next.js frontend
```

## Skills (Pipeline Stages)
| Command | Skill | Description |
|---------|-------|-------------|
| `/kd-brainstorm` | `kd-brainstorm` | Fact Ledger → parallel research → brainstorm → draft spec |
| `/kd-handoff-spec` | `kd-handoff-spec` | Guardrail validation → phase decomposition → contract handoff |
| `/kd-dev` | `kd-dev` | Pick up handoff → implement current phase → self-verify |
| `/kd-qa` | `kd-qa` | Run checks → Progress Ledger → loop detection → route |
| `/kd-handoff-dev` | `kd-handoff-dev` | Finalize → prepare release handoff |
| `/kd-release` | `kd-release` | Deploy → verify → content handoff |
| `/kd-content` | `kd-content` | Generate changelog, blog, docs |

## Core Principles
1. **Fact-first** — Gather and verify facts before brainstorming solutions (Magentic-One Task Ledger pattern)
2. **Phased execution** — Break work into independently testable phases (required for effort ≥ M)
3. **Contract-driven handoffs** — Every handoff includes task_description, acceptance_criteria, context_keys, output_mode
4. **Guardrailed specs** — Automated validation before specs enter the dev pipeline (CrewAI guardrail pattern)
5. **Loop detection** — Progress Ledger prevents infinite QA→dev cycles; escalate at loop_count ≥ 3
6. **Self-improvement** — Every user correction becomes a lesson in `_context/lessons.md`; review at session start

## Agent Behavior Rules

These rules govern how every agent operates, regardless of pipeline stage.

### Planning
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions).
- If something goes sideways, **STOP and re-plan** — never push a failing approach.
- Use plan mode for verification steps, not just implementation.

### Subagent Delegation
- Use subagents to keep the main context window clean.
- Offload research, exploration, and parallel analysis to subagents.
- One task per subagent for focused execution.
- For complex problems, increase parallelism via more subagents.

### Self-Improvement Loop
- After ANY correction from the user → update `_context/lessons.md` with the pattern.
- Write rules that prevent the same mistake from recurring.
- Review `_context/lessons.md` at the start of every session.

### Verification Standard
- Never mark a task complete without proving it works.
- Ask: *"Would a staff engineer approve this?"*
- Run tests, check logs, demonstrate correctness.

### Elegance Gate
- For non-trivial changes: pause and ask *"Is there a more elegant way?"*
- If a fix feels hacky: re-implement the clean solution before submitting.
- Skip this for simple, obvious fixes — do not over-engineer.

### Autonomous Bug Fixing
- When given a bug report: fix it. Do not ask for hand-holding.
- Point at logs, errors, failing tests — then resolve them.
- Use the **Bug Fix Fast Path** (skip brainstorm/spec, go straight to `/kd-dev`).

### Simplicity & Minimal Impact
- Make every change as simple as possible. Touch minimal code.
- Find root causes. No temporary fixes. Senior developer standards.

## Pipeline Modes

### Feature Pipeline (default)
```
/kd-brainstorm → /kd-handoff-spec → /kd-dev → /kd-qa → /kd-handoff-dev → /kd-release → /kd-content
```

### Bug Fix Fast Path
```
Bug report → /kd-dev → /kd-qa → /kd-handoff-dev → /kd-release
```
- Only for clear bugs with reproducible symptoms.
- Dev agent creates a minimal handoff ticket directly (no spec required).
- Must still pass full QA gates.
- If the bug reveals a deeper architectural issue → escalate to Feature Pipeline.

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

## Context Rules
1. All agents read from `_context/` for shared state
2. All agents review `_context/lessons.md` at session start for relevant patterns
3. Work passes between agents via `_handoff/queue/` with explicit **Contract** sections
4. Completed work moves to `_handoff/archive/`
5. Specs follow lifecycle: `draft → approved → implemented → released → archived`
6. Every handoff ticket has: id, from, to, priority, status, spec, total_phases, current_phase, loop_count, output_mode
7. Handoff contracts follow rules in `_handoff/README.md`

## Quick Start
1. **Start a feature**: Tell me what you want to build → I'll load `kd-brainstorm`
2. **Report a bug**: Describe the bug → I'll use the Bug Fix Fast Path
3. **Check pipeline**: "What's in the queue?" → I'll scan `_handoff/queue/`
4. **Continue work**: "Pick up next task" → I'll find the highest-priority pending ticket

## Quality Gates
- Backend: `go vet ./...`, `go test ./...`, `go build ./...`
- Frontend: `npm run build`, `npm test`, `npm run test:e2e`
