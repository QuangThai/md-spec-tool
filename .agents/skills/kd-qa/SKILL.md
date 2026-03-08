---
name: kd-qa
description: "Run QA verification on completed dev work. Checks tests, linting, types, contracts, and acceptance criteria. Use after dev completes implementation. Triggers on: qa, verify, test, check quality."
---

# kd-qa — Quality Assurance Agent

Verify completed dev work against specs, run tests, and ensure quality standards.

## Workflow

### Step 1: Pick Up Completed Work
1. Review `_context/lessons.md` for patterns relevant to QA tasks
2. List `_handoff/queue/` for tickets where `to: dev` and `status: done` — these are dev-completed tickets ready for QA review
3. Read the handoff ticket, referenced spec, and implementation log from dev
4. Identify `current_phase` and `total_phases` from frontmatter
5. **Fail-fast**: If no completed tickets are found, STOP and report "No completed work ready for QA." If the referenced spec is missing, STOP and report the dependency.
6. **Bug Fix Fast Path**: If `spec: none`, skip spec reading. Use the ticket's Contract/Bug Details sections as the source of acceptance criteria. Still run all automated checks.

### Step 2: Automated Checks

#### Subagent Strategy (recommended)
Spawn **2 Task calls in a single message** to run backend and frontend checks in parallel:
- **Task 1**: Run all backend checks (go vet, go build, go test) and return a summary table of results
- **Task 2**: Run all frontend checks (build, test) and return a summary table of results

Each returns ONLY the results table (max 200 words). The main agent then proceeds to manual review and acceptance criteria.

If not using subagents, run checks sequentially as listed below.

#### Backend Verification
```bash
cd backend
# Vet
go vet ./...
# Build
go build ./...
# Test
go test -v ./...
```

#### Frontend Verification
```bash
cd frontend
# Build (includes typecheck)
npm run build
# Unit tests
npm test
# E2E tests (if applicable)
npm run test:e2e
```

#### Frontend React Audit (if React/Next.js files changed)
Load the `react-doctor` skill and run its full audit against the changed components. Verify performance patterns, hook usage, component structure, and Next.js best practices. Add findings to the QA Report.

### Step 3: Manual Review Checklist
For each changed file, verify:
- [ ] Code follows conventions (check `backend/AGENTS.md` or `frontend/AGENTS.md`)
- [ ] No hardcoded secrets or credentials
- [ ] Error handling is appropriate
- [ ] Types are correct and complete
- [ ] Tests cover the new functionality
- [ ] No unnecessary changes outside scope

### Step 4: Acceptance Criteria Verification
Verify acceptance criteria based on the current phase:

**Current phase criteria** (must all pass):
- [ ] Each acceptance criterion for the current phase: {PASS/FAIL — evidence}

**Regression check** (if current_phase > 1):
- [ ] Prior phase functionality still works: {PASS/FAIL — evidence}

**Full spec criteria** (only on final phase, when current_phase == total_phases):
- [ ] Each overall spec acceptance criterion: {PASS/FAIL — evidence}

**Bug fix criteria** (if `spec: none`):
- [ ] Bug symptoms no longer reproducible
- [ ] No regressions introduced (all existing tests pass)
- [ ] Root cause documented in ticket's Bug Details section

### Step 5: Generate QA Report
Create QA report in the handoff ticket:

```markdown
## QA Report
- **Date**: {ISO date}
- **Agent**: qa
- **Verdict**: PASS | FAIL | PASS-WITH-NOTES
- **Failure class**: none | code | spec | env | flake
- **Phase**: {current_phase} of {total_phases}

### Automated Checks
| Check | Result |
|-------|--------|
| Backend vet | ✅/❌ |
| Backend build | ✅/❌ |
| Backend tests | ✅/❌ ({pass}/{total}) |
| Frontend build | ✅/❌ |
| Frontend tests | ✅/❌ |
| Frontend E2E | ✅/❌ |

### Acceptance Criteria
| Criterion | Result | Evidence |
|-----------|--------|----------|
| {criterion} | ✅/❌ | {detail} |

### Issues Found
{List any issues, or "None"}

### Notes for Release
{Anything the release agent should know}
```

**Verdict definitions:**
- `PASS`: All checks green, all acceptance criteria met
- `PASS-WITH-NOTES`: All criteria met, but cosmetic observations exist (naming suggestions, minor style notes). Notes must NOT require code changes — if they do, verdict must be `FAIL`
- `FAIL`: Any check failed or acceptance criterion not met

### Step 6: Progress Ledger Check
Before routing the result, evaluate progress to detect loops and stalls.

Fill out the progress ledger:

```markdown
## Progress Ledger
- **is_request_satisfied**:
  - reason: "{Evidence-based reasoning}"
  - answer: true/false
- **is_in_loop**:
  - reason: "{Compare current issues with previous QA feedback — are the same issues recurring?}"
  - answer: true/false
- **is_progress_being_made**:
  - reason: "{Did the dev agent address the specific issues from last round?}"
  - answer: true/false
- **loop_count**: {N — number of QA→dev→QA cycles for this spec/phase}
- **instruction_or_question**:
  - reason: "{What specifically should the dev agent do differently?}"
  - answer: "{Concrete, actionable instruction — not 'fix it', but 'change X in file Y to handle case Z'}"
```

**Loop detection rules:**
- `loop_count >= 3`: **STOP.** Escalate to user with full context (all QA reports + dev logs). The problem likely requires human judgment or a spec revision.
- `is_in_loop == true` AND `loop_count >= 2`: **Re-plan.** The fix approach is wrong. Update the spec or handoff with a new technical approach before retrying.
- `is_progress_being_made == false`: Flag the specific blocker. If it's an architectural issue, escalate. If it's a misunderstanding, provide explicit code-level guidance.
- `failure_class == spec`: Route to `needs-respec` status instead of back to dev. The spec itself needs revision — code changes won't fix a spec problem.
- `failure_class == env` or `flake`: Do NOT increment `loop_count`. Log the environment issue and retry. If it persists across 2 retries, escalate as infrastructure blocker.

### Step 7: Route Result
**If PASS:**
- Print: `✅ QA passed. Run /kd-handoff-dev to prepare for release.`

**If FAIL (and loop_count < 3):**
- **Reuse the existing ticket** — do NOT create a new one:
  1. Append the QA Report and Progress Ledger to the existing ticket
  2. Reset `status: done` → `status: pending` (so dev picks it up again)
  3. Increment `loop_count` in the frontmatter
  4. The `instruction_or_question` from the Progress Ledger tells dev exactly what to fix
- Print: `❌ QA failed (loop {loop_count}). Ticket returned to dev. Run /kd-dev to fix.`

**If FAIL (and failure_class == spec):**
- Set ticket `status: needs-respec`
- Print: `🔄 Spec defect detected. Ticket marked needs-respec. Revise the spec in _context/specs/ and re-run /kd-handoff-spec.`
- Present specific spec issues with evidence

**If FAIL (and loop_count >= 3):**
- Print: `🚨 QA loop detected ({loop_count} cycles). Escalating to user.`
- Present full QA history and recommend: re-spec, pair debugging, or scope reduction

## Rules
- Always run ALL automated checks, not just the ones related to the change
- Be thorough but fair — only flag real issues
- Provide evidence for every fail
- Update `_context/metrics/` with quality data
- **Write to `_context/lessons.md`** when: (a) the same failure pattern recurs across multiple specs, (b) a QA check reveals a gap in the pipeline, or (c) `loop_count >= 2` exposes a systemic issue. Do not log routine pass/fail results.
