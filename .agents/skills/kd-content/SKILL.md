---
name: kd-content
description: "Generate content from shipped features — changelogs, blog posts, social posts, docs. Use after release to create marketing and documentation content. Triggers on: content, changelog, blog post, write about."
---

# kd-content — Content Agent

Generate content artifacts from shipped features for md-spec-tool.

## Workflow

### Step 1: Pick Up Content Ticket
1. List `_handoff/queue/` for tickets where `to: content` and `status: pending`
2. Read the content handoff ticket
3. Read the original spec for full context
4. **Fail-fast**: If no pending content tickets are found, STOP and report "No shipped features ready for content." If the referenced spec is missing, STOP and report.

### Step 2: Load Context
1. Review `_context/lessons.md` for patterns relevant to content generation
2. Read `_context/product-state.md` for product positioning
3. Read the spec's technical approach for accuracy
4. Check `_context/research/` for relevant research notes

### Step 3: Generate Content

Based on the content suggestions in the handoff, generate applicable items:

#### Subagent Strategy (recommended for multiple content types)
Spawn parallel Tasks for each content type requested in the handoff:
- **Task 1**: Generate changelog entry — return only the formatted markdown
- **Task 2**: Generate blog/social post draft — return only the formatted markdown  
- **Task 3**: Generate technical post (if applicable) — return only the formatted markdown

Each subagent receives: spec summary, feature description, target audience, key technical decisions. Each returns ONLY the final drafted content (max 800 words per piece).

The main agent reviews, edits for consistency, and saves all outputs.

#### Changelog Entry
```markdown
## [{version}] - {date}

### Added
- {Feature description from user perspective}

### Changed
- {Any changes to existing behavior}

### Fixed
- {Any bug fixes included}
```

#### Blog/Social Post Draft
```markdown
# {Catchy Title}

{Hook — why should readers care?}

## The Problem
{What pain point this solves}

## The Solution
{How md-spec-tool now handles this}

## How It Works
{Brief technical explanation — accessible to target audience}

## Try It Out
{Call to action}
```

#### Documentation Update
Draft documentation updates in `_context/content/`. If changes to `docs/` or inline code are needed, note them — a dev ticket will be created (see Scope Restriction in Step 4).

#### Technical Post (optional — when spec has interesting technical decisions)
```markdown
# Building {Feature}: {Technical Insight}

## Context
{What problem we were solving and why it was non-trivial}

## Approach
{The technical approach we chose — architecture, patterns, trade-offs}

## What We Tried
{Approaches we considered and rejected, with reasons}

## Key Implementation Details
{The interesting parts — algorithms, optimizations, integrations}
{Reference specific files/modules where relevant}

## Lessons Learned
{What we would do differently, what surprised us}
{Pull from _context/lessons.md if relevant entries exist}

## Results
{Measurable outcomes — performance improvements, security hardening, etc.}
```

### Step 4: Save Content
Save generated content to `_context/content/`:
```
_context/content/YYYY-MM-DD-{feature-slug}/
├── changelog.md
├── blog-draft.md
├── social-post.md
├── docs-update.md
└── technical-post.md  (if applicable)
```

**Scope restriction**: Content agent generates drafts in `_context/content/` ONLY. If documentation in `docs/` or inline code comments need updates, create a new dev ticket in `_handoff/queue/` instead of modifying repo files directly. Content changes must not bypass the dev/QA pipeline.

### Step 5: Archive & Complete
1. Move content handoff to `_handoff/archive/`
2. Update spec status: `released` → `archived`
3. Update `_context/product-state.md` — add `(content generated)` note to the spec's line in Recent Decisions. Do NOT modify Active Specs (that's kd-release's responsibility).

```
📝 Content generated for: {title}
📁 Output: _context/content/{slug}/
✅ Pipeline complete: brainstorm → handoff-spec → dev → qa → handoff-dev → release → content
```

## Rules
- Write for the target audience (technical vs non-technical)
- Keep content accurate — reference actual implementation
- Don't over-hype — describe what was actually built
- Tag all content with `[agent: content]`
- This is the final stage — ensure everything is archived properly
- **Drafts only**: Never modify files in `backend/` or `frontend/` directly. If repo changes are needed, create a handoff ticket for dev.
- **Fail-fast**: If required artifacts are missing, STOP and report rather than guessing.
