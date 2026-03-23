# amort — Design Document

> A continuous code improvement agent that pays down technical debt incrementally, forever.

---

## Overview

Amort is a daemon that runs against a repository, continuously exploring the codebase for technical debt opportunities and producing proposals for human review. It never implements changes. It surfaces opportunities, plans them in depth, and waits for a human to act.

The system has three components: a daemon (exploration loop + web server), a CLI client, and a web UI. One process, one binary, one `amort start`.

---

## Architecture

```
amort start
  ├── web server (:4444)
  │     ├── serves proposal UI (list, detail, approve, reject)
  │     └── REST API consumed by CLI and web UI
  │
  └── exploration loop
        ├── check: pending count < queue cap (default: 10)?
        │     no  → sleep, wait for signal or poll interval
        │     yes → run exploration cycle
        ├── store proposal in SQLite
        └── repeat
```

### Exploration cycle

Each cycle makes two `claude -p` calls:

**Call 1 — Target selection (cheap)**

The selector receives the list of pending proposal titles so it avoids overlap. It does not receive the archive — a re-proposal of something previously rejected is acceptable. The selector's job is to find what intuitively feels like the most meaningful improvement opportunity in the repo.

**Call 2 — Deep analysis (expensive)**

Deeply analyzes the selected area, reads the code carefully, understands the problem, considers alternatives, and produces a proposal with a title, summary, and full planning document.

This session is the one that gets resumed later. Its session ID is stored with the proposal. When the user runs `amort resume <id>`, they drop into this conversation with full context and can ask questions, push back, or request more detail.

### What if nothing is found?

The deep analysis agent can return a "nothing meaningful here" result. When this happens, the loop sleeps briefly before trying again. A circuit breaker prevents burning tokens on a pristine codebase — after N consecutive empty results, the sleep interval increases.

---

## Data Model

Single SQLite database at `.amort/amort.db`.

### proposals

| Column       | Type     | Description                                      |
|-------------|----------|--------------------------------------------------|
| id          | TEXT     | Claude Code session ID (also used as proposal ID)|
| status      | TEXT     | `pending`, `approved`, `rejected`                |
| title       | TEXT     | One-sentence description                         |
| summary     | TEXT     | 2-3 sentence plain-language description          |
| plan        | TEXT     | Full planning document (markdown)                |
| session_id  | TEXT     | Claude Code session ID for `--resume`            |
| created_at  | DATETIME | When the proposal was generated                  |
| resolved_at | DATETIME | When it was approved/rejected (nullable)         |

No categories. No effort estimates. You read the title and summary and decide.

---

## Queue Behavior

- Default cap: **10 pending proposals**
- The loop only runs when `pending count < cap`
- Approving or rejecting a proposal frees space and signals the loop to wake
- Wake mechanism: combination of polling (every ~30s) and direct signaling from the API handler on approve/reject
- The loop is single-threaded — one exploration at a time

---

## Interfaces

### CLI

The CLI is a thin HTTP client that talks to the daemon's REST API.

```bash
amort start                  # start daemon (loop + web server)
amort list                   # list pending proposals
amort show <id>              # show full plan for a proposal
amort approve <id>           # mark proposal as approved
amort reject <id>            # mark proposal as rejected
amort resume <id>            # runs: claude --resume <session_id>
```

`amort start` is the only command that doesn't require the daemon to be running. All other commands fail with a clear message if the daemon isn't up. Ctrl-C stops the daemon.

`amort resume <id>` is the core action. It looks up the session ID and runs `claude --resume <session_id>`, dropping you into the planning conversation. You continue where the agent left off.

### Web UI

A single-page app served by the daemon. Designed for quick scanning — works on a phone.

- Shows proposals: title, summary, plan (truncated)
- Approve and reject buttons on each card
- "Copy resume command" button → copies `claude --resume <session_id>` to clipboard
- Filter by status (pending / all)

No conversation happens in the browser. The web UI is for reading and deciding.

### REST API

```
GET    /api/proposals              # list proposals (?status=pending|approved|rejected)
GET    /api/proposals/:id          # get proposal detail
POST   /api/proposals              # create proposal
POST   /api/proposals/:id/approve  # approve
POST   /api/proposals/:id/reject   # reject
```

---

## Smell Detection Strategy

This is the core of the system and the core challenge of the project.

LLMs are not naturally great at detecting problems in existing code — they're better at generating new code than critically evaluating old code. Getting useful proposals will require iterating on heuristics, prompt design, and what context we feed the selector. The right combination is out there but it won't be obvious on day one. Expect the early prompts to produce generic suggestions ("this function is long") before they produce structural insights.

Start with the simplest version and iterate based on what the proposals actually look like.

---

## Configuration

Zero config by default. `amort start` in a git repo creates `.amort/` on first run.

```
.amort/
  amort.db          # SQLite database
```

Override defaults with flags:

```bash
amort start --port 4444      # web server port (default: 4444)
amort start --cap 5          # queue cap (default: 10)
```

`.amort/` should be added to `.gitignore`.

---

## Lifecycle

```
1. User runs `amort start` in a repo
2. Daemon starts, creates .amort/ if needed
3. Loop begins exploring, producing proposals
4. Queue fills to cap, loop goes idle
5. User reviews proposals via web UI or CLI
6. User approves/rejects, freeing queue space
7. Loop wakes, explores again
8. For approved proposals, user runs `amort resume <id>`
   → drops into Claude Code session with full planning context
   → continues conversation, refines plan
   → when satisfied, pastes plan into a fresh Claude Code session for implementation
9. Ctrl-C kills the daemon, proposals persist in SQLite
10. `amort start` again picks up where it left off
```

---

## What Amort Is Not

- It does not implement changes. Proposals are plans, not patches.
- It does not open pull requests or modify the codebase.
- It does not prioritize for you. The queue is chronological. You decide what matters.
- It does not require configuration, accounts, or external services.
- It is not a linter. It finds structural and architectural opportunities that rules engines miss.

---

## Open Questions

- **Rejection memory.** When a proposal is rejected, the loop may eventually re-propose something similar. Maintaining some form of memory from rejections (even just "don't touch this file for a while") would improve signal over time. Not critical for v1.
- **Multi-repo.** The current design assumes one repo per daemon instance. Supporting multiple repos would mean either multiple daemons or a more complex orchestration layer. Punt for now.
- **Cost control.** Each exploration cycle is a full Claude session. The queue cap provides natural throttling, but there's no explicit budget control. Could add a daily token/cost cap later.
- **Freshness.** Proposals can go stale when the target code changes. Not addressing this in v1 — the user will notice when they resume the session.
