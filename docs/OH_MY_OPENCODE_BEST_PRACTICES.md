# Oh My OpenCode — Best Practices

Short, practical guidance for using **oh-my-opencode** effectively.

## 1) Choose the right workflow

**Quick execution (default):**
- Use **Sisyphus** for day‑to‑day tasks.
- Add `ulw` / `ultrawork` in your prompt for full automation (explore → research → implement → verify).

**Planned execution (complex/critical):**
- Press **Tab** to enter **Prometheus** (Planner) mode.
- Let it interview you, then review the generated plan in `.sisyphus/plans/*.md`.
- Run **`/start-work`** so **Atlas** (orchestrator) executes the plan.

> **Rule:** Only use **Atlas** after a Prometheus plan exists.

## 2) Understand the core roles

- **Sisyphus**: Primary “builder” agent (implementation, debugging, refactoring).
- **Prometheus**: Planner that produces a detailed, verifiable plan.
- **Atlas**: Orchestrator that executes Prometheus plans via `/start-work`.
- **Librarian**: External docs and OSS example search.
- **Explore**: Fast internal codebase search.
- **Oracle**: High‑IQ consultant for architecture or tough debugging.

## 3) Provider priorities matter

Oh My OpenCode prefers providers in this order:

**Native (anthropic/openai/google) → Copilot → OpenCode Zen → Z.ai**

If you want the best experience, use **Claude Opus 4.5** for Sisyphus. Other models work but degrade orchestration quality.

## 4) Keep prompts actionable

Do:
- Provide clear acceptance criteria (e.g., “Add endpoint X returning Y; tests pass”).
- Call out constraints (performance, security, time bounds, no refactors).
- For multi‑step tasks, choose Prometheus → `/start-work`.

Don’t:
- Ask for “just fix everything” without scope.
- Mix unrelated changes in a single request.

## 5) Use ultrawork deliberately

`ulw` is great for:
- implementing a self‑contained feature
- refactoring with clear boundaries
- fixing a specific bug with a clear reproduction

Avoid `ulw` for:
- ambiguous requirements
- architectural changes without decisions recorded

## 6) Keep config minimal unless needed

The default config is good. Only override agent models if you have a strong reason.
When you do override, prefer changing **one agent at a time** and verify behavior.

## 7) Verification is part of the task

Always request or expect:
- LSP diagnostics on touched files
- Tests/builds when applicable

If tests are unavailable, at least request diagnostics and a short validation checklist.

## 8) When to use Oracle

Consult **Oracle** when:
- You’ve failed 2+ attempts at a fix
- The change spans multiple subsystems
- There are non‑obvious tradeoffs (security/perf/compat)

## 9) Common workflows (examples)

**Quick fix:**
```
ulw fix the null pointer in user handler, add a regression test
```

**Planned change:**
1. Tab → Prometheus interview
2. Review `.sisyphus/plans/*.md`
3. `/start-work`

---

If you’re not sure, start with Prometheus. The plan provides guardrails, and Atlas enforces completion.
