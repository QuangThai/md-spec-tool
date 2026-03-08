---
name: kd-handoff-spec
description: "Convert approved brainstorm specs into dev-ready handoff tickets. Use after brainstorm approval to queue work for dev agents. Triggers on: handoff spec, approve spec, queue for dev."
---

# kd-handoff-spec — Spec → Dev Handoff

Convert approved specs from brainstorm into actionable dev handoff tickets.

## Workflow

### Step 1: Validate Spec
1. Review `_context/lessons.md` for patterns relevant to spec validation
2. Read the approved spec from `_context/specs/SPEC-XXX-*.md`
3. Verify it has: Problem, Solution, Technical Approach, Acceptance Criteria
4. Scan `_handoff/queue/` and `_handoff/archive/` filenames to determine the next HO-ID (see `_handoff/README.md` § ID Allocation)
5. If incomplete, list what's missing and ask user to resolve
6. **Fail-fast**: If the spec is missing, has no Problem section, or has `status` other than `draft`, STOP and report the issue. Do not proceed with an invalid spec.

### Step 2: Guardrail Validation
Run automated quality checks before proceeding. If any check fails, return to brainstorm with specific issues. Max retries: 3.

**Programmatic checks:**
- [ ] Has Problem section (≥50 words, describes user pain)
- [ ] Has Technical Approach with Backend AND/OR Frontend subsections
- [ ] Has ≥3 Acceptance Criteria (testable, specific, measurable)
- [ ] Has Phase breakdown (required if effort ≥ M)
- [ ] No Open Questions marked as "blocker" or unresolved
- [ ] Effort estimate is proportional to scope (S: <5 files, M: 5-15 files, L: 15+ files)
- [ ] Fact Ledger present with no unresolved "Facts to Look Up"
- [ ] **Has Research Summary section** with: track statuses, tools used, and sources listed (specs without research are rejected — return to brainstorm with instruction to execute Step 3)

**LLM guardrail check:**
Evaluate the spec against this criterion:
> "Is this spec actionable for a developer with NO prior context? Reject if it contains ambiguous requirements like 'improve performance' without specific metrics, or 'refactor' without listing target files."

**If guardrail fails:** Return specific failure reasons. Do not proceed until resolved.

### Step 3: Phase Decomposition
Break the spec into implementation phases. This is **required** for effort ≥ M, recommended for effort S.

Decompose based on:
1. **Dependency order** — Phase N must not depend on Phase N+1
2. **Risk mitigation** — High-risk / high-uncertainty items first
3. **Incremental value** — Each phase delivers independently testable value
4. **Effort balance** — Each phase fits within ~1-2 agent sessions
5. **Rollback independence** — Reverting Phase N does not break Phase N-1

Each phase must specify:
- Scope (exact files affected)
- Acceptance criteria (testable per-phase)
- Dependencies on prior phases
- Estimated effort (S/M within the phase)

**Single-phase normalization**: For effort S with no complex dependencies, create a single phase:
- Phase 1: Complete implementation (scope = all files, criteria = all acceptance criteria)
- This ensures downstream agents always have `current_phase` and per-phase criteria to work with.

**Persist phases to spec**: After decomposition, update the spec file in `_context/specs/` with the finalized phase breakdown. The spec is the durable source of truth; the handoff ticket references it. If phases change during dev/QA, the spec must be updated too.

### Step 4: Update Spec Status
Edit the spec file — change `status: draft` → `status: approved`

### Step 5: Create Handoff Ticket
Create a handoff file in `_handoff/queue/`:

```markdown
---
id: HO-{next_id}  # (determined in Step 1)
from: brainstorm
to: dev
priority: {from spec}
status: pending
created: {ISO timestamp}
spec: SPEC-XXX
total_phases: {N}  # {N — minimum 1, even for single-phase work}
current_phase: 1
loop_count: 0
output_mode: full_history
---

# {Spec Title}

## Contract
- **task_description**: {What the dev agent should do — full context, no ambiguity}
- **acceptance_criteria**: {What "done" looks like — copied from spec, refined for testability}
- **context_keys**: {Which `_context/` artifacts to read — specs, research, decisions}
- **output_mode**: full_history

## Context
{Brief summary of the problem and approved solution}

## Scope
- **Backend**: {list affected areas — endpoints, models, services, migrations}
- **Frontend**: {list affected areas — components, routes, hooks, API calls}
- **Shared**: {any _context updates needed}

## Implementation Plan (Phased)

### Phase 1: {Phase Name} — {priority}
- **Scope**: {exact files to modify}
- **Tasks**:
  1. {Step-by-step implementation order}
  2. {Dependencies between steps}
- **Acceptance Criteria**:
  - [ ] {Testable criterion}
- **Depends on**: none

### Phase 2: {Phase Name} — {priority}
- **Scope**: {exact files to modify}
- **Tasks**:
  1. {Steps}
- **Acceptance Criteria**:
  - [ ] {Testable criterion}
- **Depends on**: Phase 1

## Deliverables
- [ ] Backend implementation (per phase)
- [ ] Frontend implementation (per phase)
- [ ] Tests (unit + integration, per phase)
- [ ] AGENTS.md updates if new conventions/patterns

## Dev Notes
- Relevant files: {list key files to modify}
- Related tests: {list test files}
- Migration needed: {yes/no}

## References
- Spec: `_context/specs/SPEC-XXX-*.md`
- Research: `_context/research/YYYY-MM-DD-*.md`
- Conventions: `backend/AGENTS.md`
- Conventions: `frontend/AGENTS.md`
```

### Step 6: Update Product State
Append to `_context/product-state.md`:
- Add spec to "Active Specs" section
- Add decision summary to "Recent Decisions" section

### Step 7: Notify
Print summary:
```
✅ Handoff created: _handoff/queue/{filename}
📋 Spec: _context/specs/SPEC-XXX-*.md (approved)
🎯 Priority: {priority}
⏭️ Next: Run /kd-dev to start implementation
```

## Rules
- Never create handoff without an approved spec
- Always include file paths for dev to modify
- Always reference relevant AGENTS.md conventions
