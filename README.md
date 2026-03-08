# md-spec-tool — Agent-Driven Development Pipeline

> A Markdown Spec Tool that converts Excel/TSV specs to Markdown with AI suggestions, powered by a local-first, file-based workflow system that orchestrates AI agents across the full product lifecycle — from brainstorming to content generation — using shared context and structured handoffs.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Pipeline Stages](#pipeline-stages)
- [Getting Started](#getting-started)
- [Shared Context Layer](#shared-context-layer)
- [Handoff System](#handoff-system)
- [Usage Guide](#usage-guide)
- [Spec Lifecycle](#spec-lifecycle)
- [Error Handling & Feedback Loops](#error-handling--feedback-loops)
- [Project Structure](#project-structure)

---

## Overview

md-spec-tool uses a **7-stage agent pipeline** to manage the entire product development lifecycle. Each stage is implemented as an [Amp skill](https://ampcode.com) that reads from a shared context layer and passes structured work items through a file-based handoff queue.

**Key design principles:**

- **Local-first** — All state lives in the filesystem. No external services required.
- **Context-sharing via symlinks** — The `_context/` folder is symlinked between the product decisions repo and the codebase, giving all agents a unified view of product state.
- **Structured handoffs** — Work moves between agents through markdown files with YAML frontmatter in `_handoff/queue/`, ensuring nothing is lost between sessions.
- **Human-in-the-loop** — Every critical transition (spec approval, deployment) requires explicit user confirmation.

---

## Architecture

### High-Level Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AGENT PIPELINE                                   │
│                                                                         │
│  ┌──────────┐    ┌──────────────┐    ┌────────┐    ┌────────┐          │
│  │ BRAINSTORM├───►│ HANDOFF-SPEC ├───►│  DEV   ├───►│   QA   │          │
│  │ Research  │    │ Approve &    │    │ Build  │    │ Verify │          │
│  │ & Ideate  │    │ Queue        │    │ & Test │    │ & Gate │          │
│  └──────────┘    └──────────────┘    └────────┘    └───┬────┘          │
│                                                        │                │
│                                          FAIL ◄────────┤                │
│                                          (loop back)   │ PASS           │
│                                                        ▼                │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐              │
│  │ CONTENT  │◄───┤   RELEASE    │◄───┤  HANDOFF-DEV     │              │
│  │ Blog,    │    │ Deploy &     │    │  Finalize &      │              │
│  │ Changelog│    │ Verify       │    │  Prepare Release │              │
│  └──────────┘    └──────────────┘    └──────────────────┘              │
└─────────────────────────────────────────────────────────────────────────┘
         ▲                    ▲                    ▲
         │                    │                    │
    ┌────┴────────────────────┴────────────────────┴────┐
    │              _context/ (Shared State)              │
    │  product-state.md │ specs/ │ decisions/ │ metrics/ │
    └───────────────────────────────────────────────────┘
```

### Data Flow

```
  User idea
    │
    ▼
  _context/specs/SPEC-XXX.md          ← brainstorm creates draft
    │
    ▼
  _handoff/queue/HO-XXX.md           ← handoff-spec creates ticket
    │
    ▼
  Codebase changes                    ← dev implements
    │
    ▼
  QA Report (in handoff ticket)       ← qa verifies
    │
    ▼
  _handoff/queue/HO-XXX-release.md   ← handoff-dev creates release ticket
    │
    ▼
  Production deployment               ← release deploys
    │
    ▼
  _context/content/YYYY-MM-DD-slug/  ← content generates artifacts
    │
    ▼
  _handoff/archive/                   ← everything archived
```

---

## Pipeline Stages


| #   | Stage            | Command            | Agent Role                                  | Input                | Output                              |
| --- | ---------------- | ------------------ | ------------------------------------------- | -------------------- | ----------------------------------- |
| 1   | **Discovery**    | `/kd-brainstorm`   | Research & ideate solutions                 | User idea or problem | Draft spec in `_context/specs/`     |
| 2   | **Approval**     | `/kd-handoff-spec` | Validate & queue for dev                    | Approved spec        | Handoff ticket in `_handoff/queue/` |
| 3   | **Development**  | `/kd-dev`          | Implement across backend & frontend         | Handoff ticket       | Code changes + tests                |
| 4   | **Quality**      | `/kd-qa`           | Run tests, lint, verify acceptance criteria | Completed dev work   | QA report (PASS/FAIL)               |
| 5   | **Finalization** | `/kd-handoff-dev`  | Prepare release package                     | QA-passed ticket     | Release handoff ticket              |
| 6   | **Release**      | `/kd-release`      | Deploy & verify production                  | Release ticket       | Live deployment                     |
| 7   | **Content**      | `/kd-content`      | Generate changelog, blog, docs              | Content ticket       | Content artifacts                   |


---

## Getting Started

### Prerequisites

- [Amp](https://ampcode.com) installed and configured
- Workspace with both repos cloned:
  ```
  Workspace/
  ├── backend/    # Go 1.24 + Gin + OpenAI + SQLite
  └── frontend/   # Next.js 16 + React 19
  ```

### Quick Start

**1. Start a new feature:**

```
/kd-brainstorm
> "I want to add real-time collaboration to the analysis page"
```

**2. Check what's in the pipeline:**

```
> "What's in the queue?"
```

The agent scans `_handoff/queue/` and reports pending work by priority.

**3. Continue work on the next task:**

```
> "Pick up next task."
```

The agent picks the highest-priority pending ticket and loads the appropriate skill.

**4. Run the full pipeline sequentially:**

```
/kd-brainstorm     →  Draft & approve a spec
/kd-handoff-spec   →  Queue it for development
/kd-dev            →  Implement the feature
/kd-qa             →  Verify quality
/kd-handoff-dev    →  Finalize for release
/kd-release        →  Deploy to production
/kd-content        →  Generate content artifacts
```

---

## Shared Context Layer

The `_context/` directory is the **single source of truth** shared between all agents. It is designed to be symlinked between the product decisions repo and the codebase repos so that both Research Agents and Dev Agents operate on an identical state.

```
_context/
├── product-state.md              # Current priorities, active specs, quality metrics
├── specs/                        # Feature specifications
│   └── SPEC-XXX-feature-name.md  #   Lifecycle: draft → approved → implemented → released → archived
├── decisions/                    # Architecture & product decision records
│   └── YYYY-MM-DD-decision.md    #   Immutable once recorded
├── research/                     # Research notes, competitive analysis
│   └── YYYY-MM-DD-topic.md       #   Created by brainstorm agent
├── design/                       # Design docs, wireframes, UX decisions
│   └── DESIGN-XXX-title.md       #   Referenced by specs
├── metrics/                      # Quality metrics, KPIs, benchmarks
│   └── current-metrics.md        #   Updated by QA agent
└── content/                      # Generated content artifacts
    └── YYYY-MM-DD-feature-slug/  #   Created by content agent
        ├── changelog.md
        ├── blog-draft.md
        └── social-post.md
```

### Symlink Setup

To share context between repos, create a symlink from each repo's root:

```bash
# On Linux/macOS
ln -s /path/to/Workspace/_context /path/to/docs-repo/_context
ln -s /path/to/Workspace/_context /path/to/codebase-repo/_context

# On Windows (run as Administrator)
mklink /D "D:\docs-repo\_context" "D:\Workspace\_context"
mklink /D "D:\codebase-repo\_context" "D:\Workspace\_context"
```

### Context Rules

1. **Append-only** — Documents are never deleted; they move through lifecycle states and eventually get archived.
2. **Timestamped** — Every entry carries an ISO 8601 date for traceability.
3. **Agent-tagged** — Every entry records which agent created or modified it (e.g., `[agent: brainstorm]`).
4. **Status-tracked** — Specs follow a strict lifecycle: `draft → approved → implemented → released → archived`.

---

## Handoff System

The `_handoff/` directory is the **inter-agent work queue**. It uses structured markdown files with YAML frontmatter to pass work between pipeline stages.

```
_handoff/
├── queue/      # Active work items awaiting pickup
│   └── HO-XXX-{from}-to-{to}-{title}.md
└── archive/    # Completed work items (audit trail)
    └── (same naming convention)
```

### Handoff Ticket Structure

Every handoff ticket follows this schema:

```yaml
---
id: HO-001                           # Unique identifier
from: brainstorm                      # Originating stage
to: dev                               # Target stage
priority: P1                          # P0 (critical) | P1 (high) | P2 (normal)
status: pending                       # pending | in-progress | done | blocked
created: 2026-03-06T14:30            # ISO 8601 timestamp
spec: SPEC-001                        # Reference to source spec
---
```

### Routing Rules


| From         | To        | Trigger                                                    |
| ------------ | --------- | ---------------------------------------------------------- |
| `brainstorm` | `dev`     | Spec approved via `/kd-handoff-spec`                       |
| `dev`        | `qa`      | Implementation marked done (implicit — same ticket)        |
| `qa`         | `dev`     | QA fails — feedback loop with specific issues              |
| `dev`        | `release` | QA passes → `/kd-handoff-dev` creates release ticket       |
| `release`    | `content` | Deployment verified → `/kd-release` creates content ticket |


---

## Usage Guide

### Stage 1 — Brainstorm (`/kd-brainstorm`)

Start a product discovery session. The agent will:

1. Load current product state and past decisions from `_context/`
2. Explore the problem space with 2-3 solution approaches
3. Run an Exa MCP best-practice scan and synthesize actionable guidance
4. Produce a draft spec in `_context/specs/SPEC-XXX-title.md`
5. Present the draft for your review

```
/kd-brainstorm
> "We need batch export — users want to download all analyses at once"
```

**Output:** Draft spec with problem statement, proposed solution, technical approach, acceptance criteria, and linked research notes.

**Next step:** Review the spec. When satisfied, run `/kd-handoff-spec`.

---

### Stage 2 — Handoff Spec (`/kd-handoff-spec`)

Convert an approved spec into a dev-ready handoff ticket. The agent will:

1. Validate the spec has all required sections
2. Update spec status from `draft` → `approved`
3. Create a structured handoff ticket in `_handoff/queue/`
4. Update `_context/product-state.md` with the active spec

```
/kd-handoff-spec
```

**Output:** Handoff ticket with implementation plan, file paths, and acceptance criteria.

**Next step:** Run `/kd-dev` to begin implementation.

---

### Stage 3 — Development (`/kd-dev`)

Pick up the highest-priority handoff ticket and implement. The agent will:

1. Scan `_handoff/queue/` for pending `to: dev` tickets, sorted by priority
2. Load the spec and AGENTS.md conventions
3. Implement across backend (Go/Gin) and frontend (Next.js) as needed
4. Run self-verification: vet, build, test, and lint
5. Update the handoff ticket with an implementation log

```
/kd-dev
```

**Backend conventions enforced:**

- Go 1.24+, Gin handlers, structured packages under `internal/`
- `go vet ./...` + `go test ./...` + `go build ./...`

**Frontend conventions enforced:**

- TypeScript strict mode, `"use client"` directives
- `npm run build` + `npm test`

**Next step:** Run `/kd-qa` to verify the implementation.

---

### Stage 4 — Quality Assurance (`/kd-qa`)

Run comprehensive verification against the completed work. The agent will:

1. Execute all automated checks (lint, types, tests, contract verification)
2. Review changed files against AGENTS.md conventions
3. Verify each acceptance criterion from the spec
4. Generate a structured QA report with PASS/FAIL verdict

```
/kd-qa
```

**On PASS:** Move forward with `/kd-handoff-dev`.

**On FAIL:** A feedback ticket is automatically queued back to dev with specific issues. Run `/kd-dev` to address them.

---

### Stage 5 — Finalize (`/kd-handoff-dev`)

Prepare QA-passed work for release. The agent will:

1. Verify the QA report shows PASS
2. Ensure AGENTS.md documentation is updated
3. Update spec status from `approved` → `implemented`
4. Create a release handoff ticket with deploy notes and rollback plan
5. Archive the original dev handoff ticket

```
/kd-handoff-dev
```

**Next step:** Run `/kd-release` to deploy.

---

### Stage 6 — Release (`/kd-release`)

Deploy to production and verify. The agent will:

1. Run the pre-deploy checklist (all quality gates)
2. Present the deploy command for user approval — **never auto-deploys**
3. Guide post-deploy verification (health checks, smoke tests)
4. Update product state and archive the release ticket
5. Create a content handoff ticket for the shipped feature

```
/kd-release
```

**Deploy script reference:** `docs/deploy.sh`

```bash
bash docs/deploy.sh develop
```

**Next step:** Run `/kd-content` to generate content.

---

### Stage 7 — Content (`/kd-content`)

Generate content artifacts for the shipped feature. The agent will:

1. Read the content handoff ticket and original spec
2. Generate applicable content: changelog, blog post, social post, documentation updates
3. Save all artifacts to `_context/content/YYYY-MM-DD-feature-slug/`
4. Archive the content ticket and mark the spec as `archived`

```
/kd-content
```

**Output directory:**

```
_context/content/2026-03-06-batch-export/
├── changelog.md
├── blog-draft.md
├── social-post.md
└── docs-update.md
```

**Pipeline complete.** 🎉

---

## Spec Lifecycle

Every feature spec passes through a strict lifecycle, tracked in its YAML frontmatter:

```
  draft ──────► approved ──────► implemented ──────► released ──────► archived
    │               │                 │                  │                │
    │               │                 │                  │                │
 brainstorm    handoff-spec         dev/qa          release           content
 creates it    approves it      code is done     deployed live     content done
```


| Status        | Set By             | Meaning                                    |
| ------------- | ------------------ | ------------------------------------------ |
| `draft`       | `/kd-brainstorm`   | Spec created, under review                 |
| `approved`    | `/kd-handoff-spec` | Spec approved, dev ticket queued           |
| `implemented` | `/kd-handoff-dev`  | Code complete, QA passed, ready for deploy |
| `released`    | `/kd-release`      | Live in production                         |
| `archived`    | `/kd-content`      | Content generated, lifecycle complete      |


---

## Error Handling & Feedback Loops

The pipeline includes a built-in feedback loop at the QA stage:

```
                    ┌──────────────────────────────┐
                    │                              │
                    ▼                              │
              ┌──────────┐     FAIL          ┌─────────┐
     ────────►│   DEV    ├──────────────────►│   QA    │
              │ Implement│                   │ Verify  │
              └──────────┘◄──────────────────┤         │
                    ▲           feedback      └────┬────┘
                    │           ticket             │
                    │                              │ PASS
                    │                              ▼
                    │                        ┌──────────┐
                    └────────────────────────┤HANDOFF-DEV│
                        (if issues found     └──────────┘
                         during release)
```

**QA Failure Flow:**

1. QA agent identifies specific issues with evidence
2. A feedback handoff ticket is created: `from: qa`, `to: dev`
3. Dev agent picks up the feedback ticket on the next `/kd-dev` run
4. Cycle repeats until QA passes

**Release Rollback Flow:**

1. If post-deploy verification fails, the release ticket contains rollback instructions
2. User executes rollback manually
3. Issues are fed back to dev via a new handoff ticket

---

## Project Structure

```
Workspace/
│
├── _context/                          # Shared context layer
│   ├── product-state.md              #   Current product state & priorities
│   ├── specs/                        #   Feature specifications (lifecycle-tracked)
│   ├── decisions/                    #   Architecture decision records
│   ├── research/                     #   Research notes & analysis
│   ├── design/                       #   Design docs & UX decisions
│   ├── metrics/                      #   Quality metrics & KPIs
│   └── content/                      #   Generated content artifacts
│
├── _handoff/                          # Inter-agent work queue
│   ├── queue/                        #   Active tickets awaiting pickup
│   └── archive/                      #   Completed tickets (audit trail)
│
├── .agents/                           # Amp agent configuration
│   └── skills/                       #   Pipeline skill definitions
│       ├── kd-brainstorm/            #     Stage 1: Discovery
│       ├── kd-handoff-spec/          #     Stage 2: Approval
│       ├── kd-dev/                   #     Stage 3: Development
│       ├── kd-qa/                    #     Stage 4: Quality Assurance
│       ├── kd-handoff-dev/           #     Stage 5: Finalization
│       ├── kd-release/              #     Stage 6: Release
│       └── kd-content/              #     Stage 7: Content
│
├── docs/                              # Operational documentation
│   └── deploy.sh                    #   Production deploy script
│
├── backend/                           # Go + Gin backend
│   ├── cmd/                          #   Entry points
│   │   ├── server/                   #     HTTP server
│   │   ├── cli/                      #     CLI tool
│   │   ├── evalrunner/               #     Eval runner
│   │   ├── usecases/                 #     Use case runner
│   │   └── usecases_diff/            #     Use case diff tool
│   ├── internal/                     #   Internal packages
│   │   ├── ai/                       #     OpenAI integration
│   │   ├── config/                   #     Configuration
│   │   ├── converter/                #     Excel/TSV → Markdown conversion
│   │   ├── diff/                     #     Diff utilities
│   │   ├── feedback/                 #     Feedback handling
│   │   ├── gsheetutils/              #     Google Sheets utilities
│   │   ├── http/                     #     HTTP handlers (Gin)
│   │   ├── quota/                    #     Quota management
│   │   ├── share/                    #     Sharing functionality
│   │   └── suggest/                  #     AI suggestions
│   ├── go.mod                        #   Go module definition
│   ├── Dockerfile                    #   Container build
│   └── AGENTS.md                     #   Backend agent conventions
│
├── frontend/                          # Next.js 16 + React 19 frontend
│   ├── app/                          #   App Router pages
│   ├── components/                   #   UI components
│   ├── hooks/                        #   Custom React hooks
│   ├── lib/                          #   Utilities & API clients
│   ├── constants/                    #   App constants
│   ├── styles/                       #   CSS & Tailwind styles
│   ├── package.json                  #   Dependencies
│   ├── Dockerfile                    #   Container build
│   └── AGENTS.md                     #   Frontend agent conventions
│
├── use-cases/                         # Test data & use case files
│
├── AGENTS.md                          # Root orchestrator configuration
└── README.md                          # ← You are here
```

---

## License

Internal project — md-spec-tool.
