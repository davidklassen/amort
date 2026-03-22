package explore

import (
	"context"
	"log/slog"
	"time"

	"github.com/davidklassen/amort/store"
)

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
		count, err := l.store.PendingCount()
		if err != nil {
			slog.Error("error checking pending count", "error", err)
			if !l.sleep(ctx, 30*time.Second) {
				return
			}
			continue
		}

		if count >= l.qcap {
			slog.Info("queue full, waiting", "pending", count, "cap", l.qcap)
			if !l.wait(ctx) {
				return
			}
			continue
		}

		slog.Info("exploring", "pending", count, "cap", l.qcap)
		proposal, err := l.explore(ctx)
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

func (l *Loop) explore(ctx context.Context) (*store.Proposal, error) {
	// TODO: implement two-call claude -p cycle
	// Call 1: target selection (cheap)
	// Call 2: deep analysis (expensive)
	_ = ctx
	return nil, nil
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
