---
name: kd-brainstorm
description: "Brainstorm product ideas, research solutions, and draft specs. Use when starting new features, exploring approaches, or running product discovery sessions. Triggers on: brainstorm, ideate, explore, research."
---

# kd-brainstorm — Product Discovery & Research

Brainstorm ideas, research solutions, analyze competitors, and produce draft specs for md-spec-tool.

## ⛔ CRITICAL RULES — Read Before Proceeding

1. **DO NOT skip Step 3 (Research) unless the skip criteria in Step 3 are ALL met.** If the Fact Ledger has ANY items in "Facts to Look Up", you MUST execute research. No exceptions.
2. **DO NOT draft a spec (Step 5) without completing research first.** A spec without research is rejected by the handoff-spec guardrail.
3. **Research MUST use tool calls.** Reading the codebase with `finder`/`Grep`/`Read` and searching the web with `web_search`/`read_web_page` are MANDATORY actions, not suggestions. If you produce a spec without having called these tools (or dispatched subagents that called them), you have violated this workflow.
4. **Every spec MUST include a "Research Summary" section** listing the tools called, sources consulted, and key findings. An empty or missing Research Summary is a spec defect.

---

## Workflow

### Step 1: Load Context

Read ALL of the following files. Do not proceed until you have read them:

1. `_context/lessons.md` — scan for patterns relevant to this topic
2. `_context/product-state.md` — current priorities and active specs
3. `_context/decisions/` — past architecture and product decisions (if any exist)
4. `_context/research/` — past research notes (check for related prior research to avoid duplication)
5. `_context/specs/` — scan existing spec filenames to determine the **next SPEC ID** (e.g., if `SPEC-005` exists, next is `SPEC-006`)
6. `_handoff/queue/` — any feedback from QA or dev that needs addressing

### Step 2: Fact Ledger

Before brainstorming solutions, gather and classify facts. This prevents hallucinated assumptions and ensures the solution is grounded in reality.

Create a structured ledger and **display it to the user**:

```markdown
## Fact Ledger

### Given / Verified Facts
- {Known truths from codebase, product-state, existing specs}
- {Confirmed constraints — infrastructure, budget, timeline}

### Facts to Look Up
- {Best practices for this problem domain — with target sources}
- {Library/framework capabilities — versions, limitations}
- {Competitor approaches — how others solved this}

### Facts to Derive
- {Performance impact estimates — based on current metrics}
- {Effort breakdown — files affected, migration complexity}
- {Dependency analysis — what breaks if we change X}

### Educated Guesses
- {Hypotheses requiring validation — "Redis can handle N req/s"}
- {Assumptions about user behavior or system load}
- {Confidence level: High / Medium / Low}
```

Complete the ledger with what you know from Step 1.

**⛔ GATE CHECK — after completing the Fact Ledger, evaluate:**
- Count the items in "Facts to Look Up". Call this `LOOKUP_COUNT`.
- Count the items in "Educated Guesses". Call this `GUESS_COUNT`.
- If `LOOKUP_COUNT == 0` AND `GUESS_COUNT == 0` AND estimated effort is S → you MAY skip to Step 4. **Print: "RESEARCH SKIP: No unknowns, effort S. Proceeding to Step 4."**
- Otherwise → you MUST proceed to Step 3. **Print: "RESEARCH REQUIRED: {LOOKUP_COUNT} items to look up, {GUESS_COUNT} guesses to validate. Proceeding to Step 3."**

**DO NOT silently skip Step 3. You MUST print one of the two messages above.**

### Step 3: Research

**⛔ THIS STEP IS MANDATORY unless the Gate Check in Step 2 explicitly printed "RESEARCH SKIP".**

**If you are reading this step, you MUST execute research using tool calls. Do not summarize what you "would" do — actually do it.**

#### 3A: Determine Research Scope

Based on the Fact Ledger, determine which research tracks are needed:

| Track | When to Run | Mandatory? |
|-------|-------------|------------|
| **Track A: Technical Feasibility** | ALWAYS | YES — every feature needs codebase + docs research |
| **Track B: Best Practices** | When "Facts to Look Up" contains best-practice or pattern questions | YES if applicable |
| **Track C: Security** | When ANY security trigger matches (see list below) | YES if triggered |

**Security triggers — run Track C if the feature involves ANY of:**
- Authentication or authorization changes
- User input handling (forms, uploads, API parameters)
- External API integration
- File upload or download
- Data persistence changes (new tables, schema changes)
- Secrets, credentials, or API keys
- Permission or access control changes
- Any endpoint exposed to the internet

#### 3B: Execute Research

You MUST execute research using one of two strategies. Choose based on your platform's capabilities:

---

**Strategy 1: Parallel Subagents (PREFERRED — use if Task tool is available)**

Spawn one Task call per track in a **single assistant message**. This runs them in parallel.

Each Task call MUST include in its prompt:
1. The research goal (copied from the track description below)
2. The specific questions to answer (copied from "Facts to Look Up" in the Fact Ledger)
3. The tools to use (listed per track below)
4. This exact output instruction: *"Return ONLY a structured summary (max 500 words). Do not include raw search results, full page contents, or intermediate reasoning. Format: bullet points grouped by finding category. Include URLs for all web sources consulted."*

---

**Strategy 2: Sequential Research (FALLBACK — use if Task tool is unavailable)**

Execute the research tracks sequentially in the main context. For each track, make the tool calls directly. Summarize findings after each track before proceeding to the next.

---

**⛔ IMPORTANT: Regardless of strategy, you MUST make actual tool calls. The following tools MUST be called (not just mentioned):**

- `finder` or `Grep` or `Read` — to search the local codebase (minimum 1 call)
- `web_search` or `read_web_page` — to search external documentation (minimum 1 call per applicable track)

**If a tool is unavailable in your environment, note it explicitly in the Research Summary and move to the next available tool. Do NOT use tool unavailability as a reason to skip all research.**

---

#### Track A — Technical Feasibility & Codebase Analysis

**Research goal**: Can we build this with our current stack? What existing code can we reuse or extend?

**MANDATORY actions (execute ALL):**

1. **Search the local codebase** — Use `finder` to find existing implementations related to the feature. Search for relevant function names, types, patterns, file paths. Record what exists.
2. **Read relevant source files** — Use `Read` to examine the most relevant files found. Understand current patterns, conventions, data structures.
3. **Check official documentation** — Use `web_search` to search official docs for the relevant library/framework (e.g., "Gin middleware authentication", "Next.js 16 server components data fetching"). Use `read_web_page` to read the most relevant doc pages.

**Expected output** (include ALL sections):
1. Existing patterns found in codebase (with file paths)
2. Official documentation guidance (with URLs)
3. Library capabilities and limitations
4. Integration risks
5. Recommended approach with specific file paths to modify

#### Track B — External Best Practices & Patterns

**Research goal**: How have others solved this? What are the proven patterns?

**MANDATORY actions (execute ALL):**

1. **Search official documentation** — Use `web_search` for authoritative library/framework docs related to the feature pattern
2. **Search for implementation patterns** — Use `web_search` for blog posts, tutorials, reference implementations (e.g., "Go Gin file upload best practices 2025", "Next.js form validation patterns")
3. **Read top results** — Use `read_web_page` on the 2-3 most relevant results to extract specific patterns and code examples

**Expected output** (include ALL sections):
1. Top 3-5 sources consulted (with URLs)
2. 5-10 distilled best practices applicable to md-spec-tool
3. Approaches considered and rejected (with reasons)
4. Clear recommendation with trade-offs

#### Track C — Security & Compliance

**⛔ MANDATORY when ANY security trigger matches. Skip ONLY when NONE match.**

When skipping, print: **"TRACK C SKIP: No security triggers matched."**

**Research goal**: What could go wrong? What must we protect against?

**MANDATORY actions (execute ALL):**

1. **Search official security docs** — Use `web_search` for framework-specific security documentation (e.g., "Gin CORS configuration", "Next.js CSRF protection")
2. **Search OWASP guidelines** — Use `web_search` for relevant OWASP guidelines (e.g., "OWASP file upload security", "OWASP input validation")
3. **Audit existing security patterns** — Use `finder` or `Grep` to review existing auth, validation, and sanitization code in the codebase

**Expected output** (include ALL sections):
1. Applicable security standards (with URLs)
2. Potential vulnerabilities identified
3. Required mitigations (specific, actionable)
4. Compliance checklist

#### 3C: Synthesis

**⛔ DO NOT proceed to Step 4 until this synthesis is complete.**

After research completes (all subagents return, or all sequential tracks finish):

1. **Combine findings** — Resolve conflicts between sources. If Track A says "use pattern X" but Track B says "pattern X is deprecated", resolve it.
2. **Update the Fact Ledger** — For each item:
   - "Facts to Look Up" → Move to "Given / Verified Facts" with source citation
   - "Educated Guesses" → Either promote to verified (with evidence) or mark as invalidated
   - If any "Facts to Look Up" remain unresolved → add them to the spec's "Open Questions"
3. **Write Research Note** — Save to `_context/research/YYYY-MM-DD-{topic}.md`:
   ```markdown
   # Research: {Topic}
   Date: {YYYY-MM-DD}
   Agent: brainstorm

   ## Sources Consulted
   - {URL or file path} — {what was found}

   ## Key Findings
   - {finding 1}
   - {finding 2}

   ## Best Practices (applicable to md-spec-tool)
   1. {practice}
   2. {practice}

   ## Recommendation
   {Clear recommendation with trade-offs}

   ## Rejected Approaches
   - {approach} — rejected because {reason}
   ```
4. **Print research completion message:**
   ```
   ✅ RESEARCH COMPLETE
   - Track A: {completed|skipped} — {N} codebase files examined, {N} doc pages read
   - Track B: {completed|skipped} — {N} sources consulted
   - Track C: {completed|skipped|not triggered} — {reason if skipped}
   - Research note saved: _context/research/{filename}
   - Fact Ledger: {N} items resolved, {N} remaining as Open Questions
   ```

### Step 4: Brainstorm Solutions

**⛔ PREREQUISITE: Step 3 must be completed (or explicitly skipped via Gate Check). If you arrived here without research and the Gate Check did not print "RESEARCH SKIP", STOP and go back to Step 3.**

Based on verified facts and research findings, explore:
- **Problem definition** — What user pain does this solve? (grounded in Fact Ledger)
- **Solution options** — At least 2-3 approaches with trade-offs (informed by Track B findings)
- **Technical feasibility** — Validated against current architecture and codebase patterns (informed by Track A findings)
- **Scope estimation** — Small / Medium / Large effort (backed by dependency analysis from codebase search)
- **Risk assessment** — What could go wrong? (informed by Track C findings, if applicable)

### Step 5: Draft Spec

Create a **draft spec** in `_context/specs/` using the next available SPEC ID (determined in Step 1):

```markdown
---
id: SPEC-{next_id}
title: Feature Title
status: draft
priority: P0|P1|P2
effort: S|M|L
created: YYYY-MM-DD
author: agent:brainstorm
---

# SPEC-{next_id}: Feature Title

## Problem
What problem does this solve? (≥50 words, describe user pain)

## Fact Ledger
### Verified Facts
- {Key facts established during research — with source citations}

### Assumptions
- {Educated guesses with confidence levels}

## Research Summary
- **Track A (Technical Feasibility)**: {1-2 sentence summary of findings}
- **Track B (Best Practices)**: {1-2 sentence summary of findings}
- **Track C (Security)**: {1-2 sentence summary, or "Not triggered — no security implications"}
- **Research note**: `_context/research/{filename}`
- **Tools used**: {list of tools actually called: finder, web_search, read_web_page, etc.}
- **Sources**: {list of URLs and file paths consulted}

## Proposed Solution
How to solve it.

## Technical Approach
### Backend Changes
- Endpoints, models, services affected

### Frontend Changes
- Components, routes, API calls affected

## Non-Goals
- {What this spec explicitly does NOT cover}
- {Scope boundaries to prevent feature creep}

## Success Metrics
- {How to measure if this feature is successful}
- {Quantitative where possible: latency, error rate, usage}

## Rollout & Observability
- {Feature flag needed: yes/no}
- {Migration/backfill required: yes/no}
- {Key metrics to monitor post-deploy}

## Phases (required for effort ≥ M)
### Phase 1: {Name}
- Scope: {files}
- Acceptance Criteria: {testable}

### Phase 2: {Name}
- Scope: {files}
- Acceptance Criteria: {testable}
- Depends on: Phase 1

## Acceptance Criteria
- [ ] Testable criterion 1
- [ ] Testable criterion 2
- [ ] Testable criterion 3

## Open Questions
- Things to resolve before dev (mark as "blocker" or "nice-to-have")
- {Include any unresolved "Facts to Look Up" from the Fact Ledger}
```

### Step 6: User Review

Present the draft to the user for review:
- Summarize key decisions and trade-offs
- Highlight open questions
- List which research sources most influenced the design
- Ask for approval or iteration

**On approval**: User says "approve" or runs `/kd-handoff-spec` to move to dev pipeline.
**On iteration**: Refine based on feedback and re-present.

**⛔ DO NOT proceed to `/kd-handoff-spec` automatically. Always wait for explicit user approval.**

---

## Rules
- Always check existing code before proposing new patterns
- Prefer extending existing architecture over introducing new dependencies
- Tag all outputs with `[agent: brainstorm]`
- Never skip the Fact Ledger — even for small features, fill in at least "Given / Verified Facts"
- If a research subagent fails or returns empty results, note the gap in the spec's "Open Questions" section rather than guessing
- **A spec without a Research Summary section is invalid** — the handoff-spec guardrail will reject it
- **"I already know the answer" is NOT a valid reason to skip research** — verify your knowledge with tool calls

## Self-Validation Checklist

Before presenting the spec to the user, verify ALL of the following. If any item is NO, fix it before proceeding:

- [ ] Fact Ledger was created and displayed to the user
- [ ] Gate Check message was printed ("RESEARCH REQUIRED" or "RESEARCH SKIP")
- [ ] If research was required: at least 1 codebase search tool was called (finder/Grep/Read)
- [ ] If research was required: at least 1 web search tool was called (web_search/read_web_page)
- [ ] Research completion message was printed with track status
- [ ] Research note was saved to `_context/research/`
- [ ] Spec includes a Research Summary section with tools used and sources listed
- [ ] All "Facts to Look Up" items are either resolved or listed in Open Questions
- [ ] Spec has ≥3 testable Acceptance Criteria
- [ ] Problem section is ≥50 words
