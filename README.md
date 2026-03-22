# amort

A continuous code improvement agent that pays down technical debt incrementally, forever.

Amort runs as a daemon against your repository, exploring the codebase for technical debt and producing proposals for human review.

## How it works

A single process runs an exploration loop and a web server. The loop examines your code, analyzes it with Claude, and produces proposals. When the pending queue fills up (default: 10), it stops and waits for you to review. Approve or reject proposals to free space and resume exploration.

Each proposal has a **title**, **summary**, and **plan**. The plan is produced by a Claude Code session that you can resume at any time to ask questions, push back, or refine.

## Install

```bash
go install github.com/davidklassen/amort@latest
```

Or build from source:

```bash
git clone https://github.com/davidklassen/amort
cd amort
make build
```

## Usage

```bash
# Start the daemon (run from a git repo)
amort start

# List pending proposals
amort list

# Show the full plan for a proposal
amort show <id>

# Approve or reject
amort approve <id>
amort reject <id>

# Resume the Claude Code planning session
amort resume <id>
```

The web UI is available at `http://localhost:4444`.

## Flags

```
--port int   daemon port (default 4444)
--cap int    max pending proposals (default 10)
```

## Development

```bash
make build     # build the binary
make test      # run tests
make lint      # run linter
make tidy      # format and tidy modules
```

Seed test data against a running daemon:

```bash
./amort start &
./scripts/seed.sh
./amort list
```
