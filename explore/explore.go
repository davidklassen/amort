package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/davidklassen/amort/store"
)

const claudeTimeout = 10 * time.Minute

type Loop struct {
	store *store.Store
	qcap  int
	wake  chan struct{}
}

func NewLoop(s *store.Store, qcap int) *Loop {
	return &Loop{
		store: s,
		qcap:  qcap,
		wake:  make(chan struct{}, 1),
	}
}

func (l *Loop) Signal() {
	select {
	case l.wake <- struct{}{}:
	default:
	}
}

func (l *Loop) Run(ctx context.Context) {
	for {
		titles, err := l.store.PendingTitles()
		if err != nil {
			slog.Error("error checking pending proposals", "error", err)
			if !l.sleep(ctx, 30*time.Second) {
				return
			}
			continue
		}

		count := len(titles)
		if count >= l.qcap {
			slog.Info("queue full, waiting", "pending", count, "cap", l.qcap)
			if !l.wait(ctx) {
				return
			}
			continue
		}

		slog.Info("exploring", "pending", count, "cap", l.qcap)
		proposal, err := l.explore(ctx, titles)
		if err != nil {
			slog.Error("exploration failed", "error", err)
			if !l.sleep(ctx, 30*time.Second) {
				return
			}
			continue
		}

		if proposal == nil {
			slog.Info("nothing found, sleeping")
			if !l.sleep(ctx, 60*time.Second) {
				return
			}
			continue
		}

		if err := l.store.Insert(proposal); err != nil {
			slog.Error("failed to store proposal", "error", err)
		} else {
			slog.Info("stored proposal", "id", proposal.ID, "title", proposal.Title)
		}
	}
}

func (l *Loop) explore(ctx context.Context, pendingTitles []string) (*store.Proposal, error) {
	prompt := `You are a senior software engineer doing a thorough code quality audit of this repository.

Explore the codebase deeply. Read files, check git history, look at module structure, trace dependencies. Use subagents to explore different modules in parallel if the codebase is large. Your goal is to find the single most meaningful improvement opportunity.

Focus on real, structural problems:
- Duplicated logic that has diverged
- Error handling gaps or silent failures
- Tight coupling between modules
- Missing abstractions that cause widespread repetition
- Fragile patterns that will break under change
- Inconsistencies that indicate copy-paste drift

Do NOT suggest trivial changes like "add comments", "add tests", "rename variables", or "improve documentation".

After thorough exploration, produce a proposal with:
- title: one sentence describing the improvement
- summary: 2-3 sentences in plain language explaining what's wrong and why it matters
- plan: a full planning document in markdown — what needs to change, why, how, alternatives considered, edge cases, risks

If after exploring you find nothing meaningful to improve, set title to empty string.`

	if len(pendingTitles) > 0 {
		var existing []string
		for _, t := range pendingTitles {
			existing = append(existing, "- "+t)
		}
		prompt += "\n\nThe following improvements have already been proposed, find something different:\n" + strings.Join(existing, "\n")
	}

	schema := `{"type":"object","properties":{"title":{"type":"string"},"summary":{"type":"string"},"plan":{"type":"string"}},"required":["title","summary","plan"]}`

	slog.Info("running exploration agent")
	result, err := runClaude(ctx, prompt, schema)
	if err != nil {
		return nil, err
	}

	var analysis struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Plan    string `json:"plan"`
	}
	if err := json.Unmarshal(result.StructuredOutput, &analysis); err != nil {
		return nil, fmt.Errorf("parse analysis: %w", err)
	}

	if analysis.Title == "" {
		return nil, nil
	}

	return &store.Proposal{
		ID:        result.SessionID,
		Title:     analysis.Title,
		Summary:   analysis.Summary,
		Plan:      analysis.Plan,
		SessionID: result.SessionID,
	}, nil
}

type claudeResult struct {
	SessionID        string          `json:"session_id"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"`
	IsError          bool            `json:"is_error"`
}

func runClaude(ctx context.Context, prompt, jsonSchema string) (*claudeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, claudeTimeout)
	defer cancel()

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--json-schema", jsonSchema,
		"--allowedTools", "Read,Glob,Grep,Bash,Agent",
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("claude command: %w", err)
	}

	var result claudeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse claude output: %w", err)
	}

	if result.IsError {
		return nil, fmt.Errorf("claude error: %s", result.Result)
	}

	return &result, nil
}

func (l *Loop) wait(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-l.wake:
		return true
	case <-time.After(30 * time.Second):
		return true
	}
}

func (l *Loop) sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
