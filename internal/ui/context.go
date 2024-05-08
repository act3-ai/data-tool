package ui

import (
	"context"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// contextKey is how we find the UI.Root function in a context.Context.
type contextKey struct{}

// NewContext returns a new Context, derived from ctx, which carries the
// provided task.
func NewContext(ctx context.Context, task *Task) context.Context {
	return context.WithValue(ctx, contextKey{}, task)
}

// FromContextOrNoop returns the current Task.
func FromContextOrNoop(ctx context.Context) *Task {
	if v := ctx.Value(contextKey{}); v != nil {
		return v.(*Task)
	}
	// log only Task
	return &Task{
		log: logger.FromContext(ctx),
	}
}
