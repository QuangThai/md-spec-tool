---
name: kd-release
description: "Execute release process — deploy, verify, and update product state. Use when ready to deploy QA-passed changes. Triggers on: release, deploy, ship."
---

# kd-release — Release Agent

Execute the release process for md-spec-tool.

## Workflow

### Step 1: Pick Up Release Ticket
1. Review `_context/lessons.md` for patterns relevant to deployment
2. List `_handoff/queue/` for tickets where `to: release` and `status: pending`
3. Read the release handoff ticket
4. Present release summary to user for final approval
5. **Fail-fast**: If no pending release tickets are found, STOP and report "No work ready for release." If the spec is missing (and `spec` is not `none`) or QA report is missing, STOP and report.

### Step 2: Pre-Deploy Checklist
Verify all quality gates:

**Backend:**
- [ ] `cd backend && go vet ./...` passes
- [ ] `cd backend && go build ./...` passes
- [ ] `cd backend && go test ./...` passes

**Frontend:**
- [ ] `cd frontend && npm run build` passes
- [ ] `cd frontend && npm test` passes

### Step 3: Deploy
Run the deploy process:
```bash
# Deploy via Docker (adjust script/commands as needed for your environment)
# 1. Build backend: cd backend && docker build -t md-spec-tool-backend .
# 2. Build frontend: cd frontend && docker build -t md-spec-tool-frontend .
# 3. Docker compose up (if applicable)
docker compose up -d --build
```

Present the deploy command to user — **do not auto-execute deploy**.

### Step 4: Post-Deploy Verification
After user confirms deploy:
- [ ] Backend health: `curl http://127.0.0.1:8000/health`
- [ ] Frontend health: `curl http://127.0.0.1:3000`
- [ ] Specific smoke tests from release ticket
- [ ] Monitor for errors (check docker logs)

### Step 5: Update Product State
1. Update `_context/product-state.md`:
   - Find the line matching the spec ID in "Active Specs" and mark it with strikethrough + `released ✅`
   - Add a summary line to "Recent Decisions" with the date and outcome
2. Update spec status: `implemented` → `released`
3. Archive release handoff to `_handoff/archive/`
4. Skip product-state.md and spec status updates for bug fixes (`spec: none`). Bug fixes are tracked only via handoff ticket history.

### Step 6: Create Content Handoff
Create content handoff in `_handoff/queue/`:

```markdown
---
id: HO-{next_id}  # (scan _handoff/queue/ and _handoff/archive/ per ID Allocation rules)
from: release
to: content
priority: {priority}
status: pending
created: {ISO timestamp}
spec: SPEC-XXX
total_phases: 1
current_phase: 1
loop_count: 0
output_mode: last_message
---

# Content: {Feature Title}

## Contract
- **task_description**: Generate content artifacts (changelog, blog, social, docs) for the shipped feature. Write for the target audience.
- **acceptance_criteria**: At least changelog entry and one blog/social draft produced. Content is accurate and references actual implementation.
- **context_keys**: _context/specs/SPEC-XXX-*.md, _context/product-state.md, _context/research/
- **output_mode**: last_message

## What Was Shipped
{User-facing summary}

## Key Features
- {Feature 1}
- {Feature 2}

## Screenshots / Demo Points
- {What to showcase}

## Technical Highlights
- {Interesting tech decisions for technical content}

## Target Audience
- {Who benefits from this feature}

## Content Suggestions
- [ ] Changelog entry
- [ ] Blog post / social post
- [ ] Documentation update
- [ ] Demo video script
```

**Bug Fix Fast Path**: Content handoff is optional for bug fixes. Only create one if the fix has user-facing impact worth communicating. Otherwise, archive the release ticket and print completion.

### Step 7: Complete
```
🚀 Released: {title}
📋 Spec: SPEC-XXX (released)
📝 Content handoff created
⏭️ Next: Run /kd-content to create content
```

## Rules
- NEVER auto-deploy — always present command and wait for user approval
- Always verify health checks post-deploy
- Always create content handoff for shipped features
- Always include rollback commands in case of issues
- **Release must verify, not fix**: All pre-deploy checks must be non-mutating. If checks fail, route back to dev — do not auto-fix during release.
- **Fail-fast**: If required artifacts are missing, STOP and report rather than guessing.
