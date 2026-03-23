---
name: commitmsg
description: Generate a concise one-liner commit message
disable-model-invocation: true
allowed-tools: Bash(git:*)
---

## Context

- Changes: !`git diff HEAD`
- Recent commits: !`git log -5 --oneline`

## Task

Output a single one-liner commit message for the changes shown above. Nothing else.

## Rules

- Output ONLY the commit message text — no explanation, no formatting, no quotes, no code blocks
- One line only, all lowercase, under 72 characters
- Match the style of recent commits shown above
- Do NOT run any git commands
- Do NOT add "Co-Authored-By" or any attribution footer
- Do NOT reference milestones, phases, or plan labels
