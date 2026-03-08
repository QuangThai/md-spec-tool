---
name: kd-handoff-dev
description: "Prepare QA-passed work for release. Finalizes implementation notes, updates docs, and queues for release. Use after QA passes. Triggers on: handoff dev, prepare release, finalize."
---

# kd-handoff-dev — Dev → Release Handoff

Finalize QA-passed work and prepare it for release.

## Workflow

### Step 1: Verify QA Status
1. Review `_context/lessons.md` for relevant patterns
2. Read handoff ticket — find the latest QA Report section and confirm verdict is `PASS` or `PASS-WITH-NOTES`. If no QA Report exists or verdict is `FAIL`, redirect to `/kd-qa`.
3. Read `current_phase` and `total_phases` from handoff ticket — this determines whether to advance to next phase or create release ticket
4. If not passed, redirect to `/kd-qa` or `/kd-dev`
5. **Fail-fast**: If no QA-passed tickets are found, STOP and report "No QA-passed work ready for handoff." If the spec referenced by the ticket is missing AND `spec` is not `none`, STOP and report. For bug fixes (`spec: none`), proceed without spec verification.

### Step 2: Phase Routing
Check if this is a multi-phase spec:
1. Read `current_phase` and `total_phases` from the handoff ticket frontmatter
2. **If `current_phase < total_phases`** (more phases remain):
   - Increment `current_phase` by 1
   - Create a new handoff ticket in `_handoff/queue/` with:
     - `from: qa`, `to: dev`
     - `current_phase: {next_phase}`
     - `loop_count: 0` (reset for new phase)
     - `origin_handoff_id: {ID of the original handoff from kd-handoff-spec}` (track lineage across phases)
     - `status: pending`
     - Updated Contract with next phase's tasks and acceptance criteria from spec
   - Archive the completed phase's handoff ticket
   - Print: `📋 Phase {current_phase}/{total_phases} complete. Next phase queued for dev.`
   - **STOP** — do not proceed to release handoff.
3. **If `current_phase == total_phases`** (final phase):
   - Continue to Step 3 (Finalize Documentation)

### Step 3: Finalize Documentation
1. Verify `AGENTS.md` updates if architecture or conventions changed:
   - Backend: `backend/AGENTS.md`
   - Frontend: `frontend/AGENTS.md`
2. Update spec status: `approved` → `implemented` (skip if `spec: none` — bug fixes have no spec to update)

### Step 4: Create Release Handoff
Create release handoff in `_handoff/queue/`:

```markdown
---
id: HO-{next_id}  # (scan _handoff/queue/ and _handoff/archive/ per ID Allocation rules)
from: dev
to: release
priority: {priority}
status: pending
created: {ISO timestamp}
spec: SPEC-XXX
total_phases: {total_phases}
current_phase: {total_phases}
loop_count: 0
origin_handoff_id: {ID of the first handoff for this spec}
output_mode: last_message
---

# Release: {Feature Title}

## Contract
- **task_description**: Deploy {Feature Title} to production. Run pre-deploy checks, present deploy command, verify post-deploy health.
- **acceptance_criteria**: All health checks pass, no errors in docker logs, smoke tests green.
- **context_keys**: {list of _context/ files — spec, QA report}
- **output_mode**: last_message

## Summary
{One-paragraph summary of what was built}

## Changes
### Backend
- {List of key changes with file paths}

### Frontend
- {List of key changes with file paths}

### Database
- Migration: {yes/no}
- Migration name: {if applicable}

## QA Status
- All automated checks: ✅
- Acceptance criteria: ✅
- QA report: {link to handoff ticket}

## Deploy Notes
- Environment variables: {any new env vars}
- Migration steps: {if applicable}
- Breaking changes: {if any}
- Rollback plan: {how to rollback if needed}

## Post-Deploy Verification
- [ ] Health check passes
- [ ] Smoke test: {specific test}
- [ ] Monitor: {what to watch}
```

**Multi-phase summary**: If `total_phases > 1`, the release handoff's "Changes" section must summarize changes from ALL phases (read archived phase tickets from `_handoff/archive/`), not just the final phase.

### Step 5: Archive Original Handoff
Move the original dev handoff ticket from `_handoff/queue/` to `_handoff/archive/`

### Step 6: Complete
```
✅ Release handoff ready: _handoff/queue/{filename}
📋 Spec: SPEC-XXX (implemented)
🚀 Next: Run /kd-release to deploy
```

## Rules
- Never skip documentation updates
- Always include rollback plan
- Always archive completed handoffs
- **Fail-fast**: If required artifacts are missing (spec, QA report, archived phase tickets), STOP and report rather than guessing
- When creating next-phase tickets, always reset `loop_count: 0` for the new phase
