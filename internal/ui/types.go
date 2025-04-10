// Package ui defines unified user interface to decouple the information source from the presentation
package ui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/act3-ai/go-common/pkg/logger"
)

// UI is a simple user interface abstraction.
type UI interface {
	// Root task
	Root(ctx context.Context) *Task

	// Run will block until the context is exceeded or Close() is called. This runs the actual UI updating the screen.
	Run(ctx context.Context) error

	// Shutdown shuts down the UI.  It returns immediately but has the side-effect that Run() should return shortly there-after.
	Shutdown()
}

// Task simply sends task updates along the provided channel.
// It also logs everything to the provided log.
// All methods are thread-safe.
type Task struct {
	name    []string
	log     *slog.Logger
	updates chan<- event
}

func newRootTask(log *slog.Logger, updates chan<- event) *Task {
	tsk := &Task{
		log:     log,
		updates: updates,
	}

	if tsk.updates != nil {
		tsk.updates <- &taskUpdate{
			eventBase: eventBase{time.Now(), tsk.name},
			done:      false,
		}
	}

	return tsk
}

// Info suggests an informational message to be displayed.
// This is a transient informational message.
func (tsk *Task) Info(a ...any) {
	msg := fmt.Sprint(a...)
	tsk.log.Info("Informational", "message", msg)
	if tsk.updates != nil {
		tsk.updates <- &infoUpdate{
			eventBase: eventBase{time.Now(), tsk.name},
			message:   msg,
		}
	}
}

// Infof suggests an informational message to be displayed.
// The message uses Printf style formatting.
// This is a transient informational message.
func (tsk *Task) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	tsk.Info(msg)
}

// SubTask returns a new nested Task where everything send is related to the parent and this child task with name.
// This is a "prefix" for all information in this task.
// You must call Complete() when done with the work in this Task.
func (tsk *Task) SubTask(name string) *Task {
	tsk.log.Info("Creating child task", "name", name)

	// we need new memory for the slice since we reuse tsk.name
	newName := make([]string, len(tsk.name)+1)
	copy(newName, tsk.name)
	newName[len(tsk.name)] = name

	newTask := &Task{
		name:    newName,
		log:     tsk.log.WithGroup(name),
		updates: tsk.updates,
	}

	if tsk.updates != nil {
		tsk.updates <- &taskUpdate{
			eventBase: eventBase{time.Now(), newTask.name},
		}
	}
	return newTask
}

// SubTaskWithProgress creates a task supporting progress.
// You must call Complete() on the progress when finished.
func (tsk *Task) SubTaskWithProgress(name string) *Progress {
	tsk.log.Info("Creating child task with progress", "name", name)
	newTask := tsk.SubTask(name)
	if tsk.updates != nil {
		tsk.updates <- &progressUpdate{
			eventBase: eventBase{time.Now(), newTask.name},
		}
	}

	return &Progress{Task: *newTask}
}

// Complete the task.  Must be called exactly once.
func (tsk *Task) Complete() {
	tsk.log.Info("Completed")
	if tsk.updates != nil {
		tsk.updates <- &taskUpdate{
			eventBase: eventBase{time.Now(), tsk.name},
			done:      true,
		}
	}
}

// Progress simply sends progress updates along the provided channel.
// It also logs to the provides `logging.LogFrontend`.
// All methods are thread-safe.
type Progress struct {
	Task
	AggregateTo *Progress
}

// SubTaskWithProgress creates a task supporting progress.
// You must call Complete() on the progress when finished.
func (p *Progress) SubTaskWithProgress(name string) *Progress {
	newProgress := p.Task.SubTaskWithProgress(name)
	newProgress.AggregateTo = p
	return newProgress
}

// Update provides a relative progress update.
func (p *Progress) Update(deltaComplete, deltaTotal int64) {
	// this logs every progress update, bumping log verbosity to 5 (highest)
	logger.V(p.log, 4).Info("Updating", "delta complete", deltaComplete, "delta total", deltaTotal) //nolint:sloglint

	if p.updates != nil {
		p.updates <- &progressUpdate{
			eventBase: eventBase{time.Now(), p.name},
			complete:  deltaComplete,
			total:     deltaTotal,
		}

		// also update aggregator
		if p.AggregateTo != nil {
			p.AggregateTo.Update(deltaComplete, deltaTotal)
		}
	}
}

// Write implements the io.Writer interface.
func (p *Progress) Write(data []byte) (int, error) {
	p.Update(int64(len(data)), 0)
	return len(data), nil
}
