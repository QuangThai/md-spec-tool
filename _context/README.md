# _context — Shared Context Layer

This folder is the **single source of truth** shared between all agents (Research, Dev, QA, Content).
It is symlinked between the `docs/` repo and the codebase repos.

## Structure

```
_context/
├── product-state.md        # Current product state, goals, priorities
├── decisions/              # Architecture & product decision records
│   └── YYYY-MM-DD-title.md
├── specs/                  # Feature specs (lifecycle-tracked)
│   └── SPEC-XXX-title.md
├── metrics/                # KPIs, benchmarks, quality metrics
│   └── current-metrics.md
├── research/               # Research notes, competitive analysis
│   └── YYYY-MM-DD-topic.md
└── design/                 # Design docs, wireframes, UX decisions
    └── DESIGN-XXX-title.md
```

## Rules
1. **Append-only** — never delete, only archive (move to `_archive/`)
2. **Timestamped** — every entry has ISO date
3. **Agent-tagged** — every entry has `[agent: brainstorm|dev|qa|content]` tag
4. **Status-tracked** — `draft → approved → implemented → released → archived`
