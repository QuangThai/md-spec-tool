# _handoff — Inter-Agent Queue

Handoff files pass work between pipeline stages.
Each file is a structured markdown document with metadata and an explicit contract.

## Flow
```
queue/          → Agent picks up → Processes → Moves to archive/
```

## Naming Convention
```
HO-XXX-{from}-to-{to}-{title}.md
```

Example: `HO-001-brainstorm-to-dev-transcript-upload-validation.md`

## Contract Chain

Every handoff is a **contract** between the producing and receiving agent. The producing agent validates its own output before handing off — never pass incomplete work downstream.

Each handoff MUST include:
- **task_description** — What the receiving agent should do (with ALL relevant context, no ambiguity)
- **acceptance_criteria** — What "done" looks like (testable, specific)
- **context_keys** — Which `_context/` artifacts the receiving agent must read
- **output_mode** — How much history to pass:
  - `full_history` — Use for backward handoffs (QA → dev) where the receiver needs to see what was tested
  - `last_message` — Use for forward handoffs (release → content) where only the final result matters

## Output Mode Guide

| Direction | Example | Mode | Rationale |
|-----------|---------|------|-----------|
| Forward (happy path) | brainstorm → dev | `full_history` | Dev needs full spec context |
| Forward (late stage) | release → content | `last_message` | Content only needs release notes |
| Backward (failure) | QA → dev | `full_history` | Dev must see what QA tested and why it failed |
| Backward (re-spec) | dev → brainstorm | `last_message` | Brainstorm needs the problem, not implementation details |

## Handoff Template
```markdown
---
id: HO-XXX
from: brainstorm|dev|qa|release
to: dev|qa|release|content
priority: P0|P1|P2
status: pending|in-progress|done|blocked|needs-respec|cancelled|release-failed
created: YYYY-MM-DDTHH:MM:SSZ
spec: SPEC-XXX (if applicable)
total_phases: {N — total implementation phases, default 1 for single-phase work}
current_phase: {N — which phase this handoff covers, default 1}
loop_count: {N — QA→dev cycle count, starts at 0}
failure_class: none|code|spec|env|flake  # (set by QA on FAIL, helps route correctly)
output_mode: full_history|last_message
---

# Title

## Contract
- **task_description**: {What the receiving agent should do — complete, unambiguous}
- **acceptance_criteria**: {What "done" looks like — testable}
- **context_keys**: {List of `_context/` files to read}

## Context
What and why.

## Deliverables
- [ ] Checklist of what needs to be done

## Acceptance Criteria
- [ ] How to verify it's done

## References
- Links to _context/ docs, PRD sections, etc.
```

## Rules
- Never create a handoff without a Contract section
- Producing agent validates own output before writing the handoff
- Backward handoffs (FAIL routing) MUST include specific, actionable instructions — not just "fix it"
- Archive completed handoffs immediately to keep the queue clean
- **Fail-fast on missing inputs**: If a required artifact (spec, ticket, config file) is missing or has malformed frontmatter, STOP immediately. Set ticket `status: blocked`, report the exact missing dependency, and do not guess forward.

## State Precedence
When states conflict between artifacts, this is the order of truth:
1. **Spec file** (`_context/specs/`) = lifecycle truth (draft → archived)
2. **Handoff ticket** (`_handoff/queue/`) = current execution truth (pending → done)
3. **product-state.md** = summary/dashboard only — derived from specs, never the source

If a ticket and spec disagree, the **ticket** is correct for execution status, the **spec** is correct for lifecycle status.

## ID Allocation
To determine the next handoff ID:
1. Scan all files in both `_handoff/queue/` and `_handoff/archive/`
2. Parse the numeric prefix from filenames (e.g., `HO-012-...` → 12)
3. Next ID = max found + 1, zero-padded to 3 digits (e.g., `HO-013`)
4. Never reuse IDs, even for archived tickets
5. Filename format: `HO-{NNN}-{from}-to-{to}-{slug}.md`

## Ticket Lifecycle
Each ticket follows a linear state machine. Only one active ticket per spec+phase at any time.

### State transitions
```
pending → in-progress → done → archived
                      ↘ blocked (escalate to user)
                      ↘ needs-respec (route back to brainstorm)
                      ↘ cancelled (user decision)
          done → release-failed (post-deploy failure → rollback + new dev ticket)
```

### Stage routing (who updates `to:`)
| Transition | Who updates | New `to` value | Notes |
|------------|-------------|----------------|-------|
| handoff-spec creates ticket | kd-handoff-spec | `dev` | Initial creation |
| Dev picks up | kd-dev | (no change) | Status: `pending` → `in-progress` |
| Dev completes | kd-dev | (no change) | Status: `in-progress` → `done` |
| QA picks up done ticket | kd-qa | (no change) | QA reads tickets where `to: dev` AND `status: done` |
| QA PASS | kd-qa | (no change) | Appends QA Report with verdict PASS |
| QA FAIL | kd-qa | (no change) | Appends QA Report, resets `status: pending`, increments `loop_count` |
| handoff-dev routes | kd-handoff-dev | `release` (or creates next phase ticket) | Archives completed ticket |
| Release completes | kd-release | `content` (new ticket) | Archives release ticket |
| Content completes | kd-content | (no change) | Archives content ticket |
| QA finds spec defect | kd-qa | (no change) | Sets `status: needs-respec`, appends reason. Route back to `/kd-brainstorm` |
| User cancels work | user | (no change) | Sets `status: cancelled`. Archive immediately. |
| Release fails | kd-release | (no change) | Sets `status: release-failed`. Creates new dev ticket with rollback instructions. |

### QA fail reuse rule
On QA failure, do NOT create a new ticket. Instead:
1. Append QA Report and Progress Ledger to the **existing** ticket
2. Reset `status: pending` (so dev picks it up again)
3. Increment `loop_count`
4. The dev agent reads the appended QA feedback on next pickup

### PASS-WITH-NOTES definition
Verdict `PASS-WITH-NOTES` means QA passed but has cosmetic or non-blocking observations. Rules:
- Only for observations that do NOT require code changes or acceptance criteria corrections
- If any note requires a code fix, the verdict must be `FAIL`
- Notes are informational for the release agent (e.g., "consider renaming this variable in a future cleanup")
