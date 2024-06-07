package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"git.act3-ace.com/ace/go-common/pkg/logger"

	"github.com/fatih/color"
	tsize "github.com/kopoli/go-terminal-size"
)

// clear the current line via special control characters.
const clearLine = "\r\033[K\r"

// complexUI is a fancy UI that simply dumps the output to the screen with prefixes.
type complexUI struct {
	// unbuffered channel, we want a crazy responsive UI, who cares about performance :)
	updates chan event

	// out is the output stream to write the presentation to for user consumption
	out *os.File
}

// NewComplexUI returns a new fancy UI that outputs messages to "out".
// Task names are prefixed to messages to provide the necessary context.
// Progress is displayed as a progress bar
//
// out must be a terminal.
func NewComplexUI(out *os.File) UI {
	return &complexUI{
		updates: make(chan event, bufferSize),
		out:     out,
	}
}

// Run implements UI.
func (u *complexUI) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithGroup("UI")

	trackers := make(map[string]*taskTracker)
	t := time.NewTicker(time.Millisecond * 200)
	var statusLine string

	// get initial term width
	size, err := tsize.GetSize()
	if err != nil {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}
	termWidth := size.Width

	// setup size listener to handle terminal resize
	sizeListener, err := tsize.NewSizeListener()
	if err != nil {
		return fmt.Errorf("unable to watch terminal size: %w", err)
	}
	defer sizeListener.Close()

	defer fmt.Fprintln(u.out)
	for {
		select {
		case size := <-sizeListener.Change:
			termWidth = size.Width
		case update, ok := <-u.updates:
			if !ok {
				return nil
			}
			str := processUpdate(log, trackers, update)

			// if length of trackers is 0, we can reset the buffer (there are no more progress messages to display)
			// optimization
			if len(trackers) == 0 {
				statusLine = ""
			}

			if str != "" {
				// clear the line and then output the informational message (that stays in the terminal)
				// redraw the progress display
				if _, err := fmt.Fprint(u.out, clearLine, str, statusLine); err != nil {
					return fmt.Errorf("unable to write to terminal: %w", err)
				}
			}
		case <-ctx.Done():
			logger.V(log, 1).InfoContext(ctx, "Context done")
			return ctx.Err()
		case <-t.C:
			// update progress message
			// only need to update if buffer has contents
			if statusLine = renderStatus(trackers, termWidth); statusLine != "" {
				// clear the line and write out the status
				if _, err := fmt.Fprint(u.out, clearLine, statusLine); err != nil {
					return err
				}
			}
		}
	}
}

// Shutdown implements UI.
func (u *complexUI) Shutdown() {
	close(u.updates)
}

// Root implements UI.
func (u *complexUI) Root(ctx context.Context) *Task {
	return newRootTask(logger.V(logger.FromContext(ctx).WithGroup("UI"), 1), u.updates)
}

// getFgColor returns a color from a list of foreground colors based on the index.
func getFgColor(i int) color.Attribute {
	colorSlice := []color.Attribute{color.FgRed, color.FgGreen, color.FgYellow, color.FgBlue, color.FgMagenta, color.FgCyan, color.FgWhite}
	return colorSlice[i%len(colorSlice)]
}

func renderStatus(trackers map[string]*taskTracker, termWidth int) string {
	const sep = " ‖ "

	// TODO We arbitrarily prune progress by the alphabetical name.
	// We should probably not show the entire level (instead of arbitrarily showing a few).

	// TODO we could also be smarter about using the long format when it fits on the terminal.
	// maybe show long format on the first few then short on the others?

	ordered := orderedTrackers(trackers)
	parts := make([]string, 0, len(ordered))
	var n int // count of real/visible characters (not control characters)

	for i, trk := range ordered {
		// show the shortened version of progress update if more than 3 trackers
		truncated := i > 3
		s := trk.counter.Format(true)
		if trk.tracker != nil {
			if s != "" {
				s += " "
			}
			s += trk.tracker.Format(truncated)
		}

		// if no update, continue
		if s == "" {
			continue
		}

		s = fmt.Sprintf("%s ↦ %s", strings.Join(trk.name, separator), s)

		newN := n + len(sep) + len(s)
		if newN >= termWidth {
			break
		}
		c := color.New(getFgColor(i))
		parts = append(parts, c.Sprint(s))
		n = newN
	}
	return strings.Join(parts, sep)
}
