# Lessons Learned — Cross-Session Agent Memory

> Review this file at the start of every session. Update it after every correction from the user.

## How to Use This File

1. **At session start**: Scan for patterns relevant to the current task.
2. **After any user correction**: Add a new entry with the pattern, the mistake, and the rule to prevent it.
3. **Periodically**: Review and consolidate — merge related lessons, remove outdated ones.

## Hygiene Rules
1. **Only log reusable, non-obvious lessons** — don't log routine implementation details.
2. **Archive stale lessons** — move lessons older than 3 months to a `## Archived Lessons` section at the bottom.
3. **Scope narrowly** — a lesson about Go error handling should say `scope: backend`, not `scope: all`.
4. **Keep an active summary** — maintain a "Top 5 Active Rules" section at the top for quick scanning.

## Format

```markdown
### [YYYY-MM-DD] Lesson Title
- **Context**: What task was being done
- **Mistake**: What went wrong
- **Correction**: What the user said
- **Rule**: The rule to follow going forward
- **Applies to**: brainstorm | dev | qa | all
- **Scope**: backend | frontend | pipeline | all
- **Severity**: critical | important | minor
- **Review after**: YYYY-MM-DD (optional — set for time-sensitive lessons)
```

---

## Lessons

<!-- Agents: append new lessons below this line -->
