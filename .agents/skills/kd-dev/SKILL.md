---
name: kd-dev
description: "Pick up handoff tickets and implement features across backend and frontend. Use when ready to code approved specs. Triggers on: dev, implement, build, code."
---

# kd-dev — Development Agent

Pick up approved handoff tickets and implement features in the md-spec-tool codebase.

## Workflow

### Bug Fix Fast Path
When invoked directly for a bug (no existing handoff ticket), create a minimal ticket before implementing:

1. Scan `_handoff/queue/` and `_handoff/archive/` to determine next HO-ID
2. Create a minimal handoff ticket in `_handoff/queue/`:

```markdown
---
id: HO-{next_id}
from: dev
to: dev
priority: {P0|P1|P2 based on severity}
status: in-progress
created: {ISO timestamp}
spec: none
total_phases: 1
current_phase: 1
loop_count: 0
output_mode: full_history
---

# Bugfix: {Brief Description}

## Contract
- **task_description**: {What the bug is and how to fix it}
- **acceptance_criteria**: {How to verify the bug is fixed}
- **context_keys**: {relevant files}

## Bug Details
- **Symptoms**: {What the user reported or what's failing}
- **Root cause**: {To be filled after investigation}
- **Files affected**: {To be filled during implementation}
```

3. Then proceed to Step 2 (Load Context) with this self-created ticket.
4. If investigation reveals a deeper architectural issue, STOP and escalate to the full Feature Pipeline (`/kd-brainstorm`).

### Step 1: Pick Up Work
1. List `_handoff/queue/` for pending tickets where `to: dev`
2. Sort by priority (P0 > P1 > P2)
3. Present to user: "Found {N} pending tickets. Starting with: {title} (P{X})"
4. Read the full handoff ticket
5. Identify current phase: `current_phase` of `total_phases` from frontmatter
6. **Fail-fast**: If no pending tickets are found and no bug was reported, STOP and report "No pending work in queue." If a ticket references a spec that doesn't exist, STOP and report the missing dependency.

### Step 2: Load Context
1. Review `_context/lessons.md` for patterns relevant to this task
2. Read the **Contract** section — this is the primary instruction set
3. Read all files listed in `context_keys`
4. Read the referenced spec from `_context/specs/` — focus on the current phase section
5. Read relevant `AGENTS.md` files for code conventions
6. Scan affected files listed in the current phase's scope using `finder` / `Read`
7. If `loop_count > 0`: Read previous QA report and Progress Ledger — address the specific `instruction_or_question`

> **Bug Fix Fast Path note**: For tickets where `spec: none`, skip steps 4–5 (spec, AGENTS.md reading) and focus on the bug details in the Contract/Bug Details sections and the affected code.

### Step 3: Implement

#### Subagent Strategy (optional, for larger changes)
For changes spanning both backend and frontend, consider spawning 2 parallel Tasks:
- **Task 1**: Backend impact analysis — scan affected files, check existing tests, identify dependencies
- **Task 2**: Frontend impact analysis — scan affected components, check API contract alignment

Each returns a concise summary (max 300 words). The main agent then implements changes directly — subagents are for analysis only, never for writing code.

#### Backend (backend/)
Follow conventions from `backend/AGENTS.md`:
- Go 1.24+, error wrapping with `fmt.Errorf`, `log/slog` logging
- Gin HTTP handlers in `internal/http/`
- AI integration in `internal/ai/`
- Domain services in `internal/` packages
- Config via `internal/config/` with `LoadConfig()`
- Tests alongside code: `go test ./...`

#### Frontend (frontend/)
Follow conventions from `frontend/AGENTS.md`:
- Next.js 16 App Router, React 19, TypeScript
- Components in `components/` with `"use client"` directive
- API clients in `lib/api.ts`
- Types in `lib/types.ts`
- Styling with Tailwind + `cn()` utility

**Before writing React code:** Load the `vercel-react-best-practices` skill for performance optimization patterns (memoization, code splitting, data fetching, bundle size). Apply its guidelines to all React/Next.js code in this implementation.

### Step 4: Self-Verify
Run checks before marking done:

**Backend:**
```bash
cd backend
go vet ./...
go test ./...
go build ./...
```

**Frontend:**
```bash
cd frontend
npm run build
npm test
```

**Frontend React Check (if React/Next.js files changed):**
Load the `react-doctor` skill and run its diagnostics against the changed components. Fix any issues it flags before proceeding.

### Step 5: Elegance Gate
Before marking done, pause and self-review:
1. Ask: *"Would a staff engineer approve this diff?"*
2. For non-trivial changes: *"Is there a more elegant way to do this?"*
3. If the fix feels hacky: re-implement the clean solution now — do not defer.
4. Skip this step for simple, obvious fixes (single-line changes, config updates).

### Step 6: Update Handoff
1. Update handoff ticket: `status: pending` → `status: in-progress` → `status: done`
2. Add implementation notes to the handoff ticket:
   ```markdown
   ## Implementation Log
   - Phase: {current_phase} of {total_phases}
   - Files changed: {list}
   - Tests added: {list}
   - Migration: {yes/no, migration name}
   - Loop iteration: {loop_count} (if retrying after QA fail)
   - QA feedback addressed: {specific items from Progress Ledger, if applicable}
   - Notes: {anything QA should know}
   ```

### Step 7: Complete
Print summary:
```
✅ Implementation done: {title} (Phase {current_phase}/{total_phases})
📁 Files changed: {count}
🧪 Tests: {pass/fail summary}
⏭️ Next: Run /kd-qa to verify
```

## Rules
- Always read existing code before modifying
- Make smallest reasonable diffs
- Follow existing patterns — don't introduce new libraries without approval
- Write tests for new functionality
- Never skip vet or type checking
- **Write to `_context/lessons.md`** when: (a) you discover a non-obvious root cause during debugging, (b) a QA loop reveals a recurring pattern, or (c) you find an undocumented codebase convention. Do not log routine implementation details.
