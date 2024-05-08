package ui

import (
	"context"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// silentUI is a UI that outputs nothing.
type silentUI struct {
	done chan struct{}
}

// NewSilentUI returns a UI that outputs nothing (quiet mode).
func NewSilentUI() UI {
	return &silentUI{
		done: make(chan struct{}),
	}
}

// Run implements UI.
func (u *silentUI) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-u.done:
		return nil
	}
}

// Shutdown implements UI.
func (u *silentUI) Shutdown() {
	close(u.done)
}

// Root implements UI.
func (u *silentUI) Root(ctx context.Context) *Task {
	return newRootTask(logger.V(logger.FromContext(ctx).WithGroup("UI"), 1), nil)
}
