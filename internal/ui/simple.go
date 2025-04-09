package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/act3-ai/go-common/pkg/logger"
)

// channel buffer size.
const bufferSize = 10

// visual separator for names.
const separator = " | "

// separator for names in the tracker map (null is not allowed in the names so this will not conflict).
const null = string(rune(0))

// simpleUI is a basic UI that simply dumps the output to the screen with prefixes.
type simpleUI struct {
	// updates is used to send all the updates along without blocking the main goroutine
	updates chan event

	// out is the output stream to write the presentation to for user consumption
	out io.Writer
}

// NewSimpleUI returns a new simple UI that simply outputs messages to "out".
// Task names are prefixed to messages to provide the necessary context.
// Progress is updated regularly.
//
// out need not be a terminal for this UI.
func NewSimpleUI(out io.Writer) UI {
	return &simpleUI{
		updates: make(chan event, bufferSize),
		out:     out,
	}
}

// Run implements UI.
func (u *simpleUI) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithGroup("UI")

	trackers := make(map[string]*taskTracker)

	t := time.NewTicker(time.Millisecond * 1000)
	for {
		select {
		case update, ok := <-u.updates:
			if !ok {
				return nil
			}
			if str := processUpdate(log, trackers, update); str != "" {
				if _, err := u.out.Write([]byte(str)); err != nil {
					return fmt.Errorf("unable to write message to output: %w", err)
				}
			}

		case <-ctx.Done():
			logger.V(log, 1).InfoContext(ctx, "Context done")
			return ctx.Err()

		case <-t.C:
			// output progress update
			// only need to update if buffer has contents
			if status := renderSimpleStatus(trackers); status != "" {
				if _, err := fmt.Fprint(u.out, status); err != nil {
					return err
				}
			}
		}
	}
}

// Shutdown implements UI.
func (u *simpleUI) Shutdown() {
	close(u.updates)
}

// Root implements UI.
func (u *simpleUI) Root(ctx context.Context) *Task {
	return newRootTask(logger.V(logger.FromContext(ctx).WithGroup("UI"), 1), u.updates)
}

func renderSimpleStatus(trackers map[string]*taskTracker) string {
	buf := &strings.Builder{}
	for i, trk := range orderedTrackers(trackers) {
		progressStr := trk.counter.Format(false)
		if trk.tracker != nil {
			if progressStr != "" {
				progressStr += " "
			}
			progressStr += trk.tracker.Format(false)
		}

		if progressStr != "" {
			prefix := strings.Join(trk.name, separator)
			if i == 0 {
				buf.WriteString("[---------- status -------\n")
			}
			if _, err := fmt.Fprintf(buf,
				"%s ↦ %s\n", prefix, progressStr); err != nil {
				panic(err)
			}
		}
	}
	if buf.Len() > 0 {
		buf.WriteString("-------------------------]\n")
	}
	return buf.String()
}
