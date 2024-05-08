package ui

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"git.act3-ace.com/ace/data/tool/internal/ui/tracker"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

func processUpdate(log *slog.Logger, trackers map[string]*taskTracker, evt event) string {
	name := evt.Name()
	prefix := strings.Join(name, null)

	log = logger.V(log, 2).With("task", name)

	trk := trackers[prefix]
	switch e := evt.(type) {
	case *infoUpdate:
		log.Info("Processing infoUpdate") //nolint:sloglint
		if trk == nil {
			panic(fmt.Sprintf("Info() called on non-existent Task %q with message %s", name, e.message))
		}
		return strings.Join(name, separator) + " ↦ " + e.message + "\n"
	case *taskUpdate:
		log.Info("Processing taskUpdate", "complete", e.done) //nolint:sloglint
		var parentTrk *taskTracker
		if len(name) > 0 {
			parent := strings.Join(name[:len(name)-1], null)
			parentTrk = trackers[parent]
		}

		if !e.done {
			// not done means we have a new Task
			// we increase the total count of the parent task if one exists

			if trk != nil {
				panic(fmt.Sprintf("Non-unique task name provided: %q", name))
			}
			trk = &taskTracker{
				name:    name,
				created: e.Time(),
				tracker: nil,
				counter: *tracker.NewCounter(),
			}
			trackers[prefix] = trk

			// if the parent task is not nil, we need to increment the total count of the parent task
			if parentTrk != nil {
				parentTrk.counter.AddTotal(1)
			}
			log.Info("Starting task", "name", strings.Join(name, separator)) //nolint:sloglint
			return ""
		}

		// the task completed
		if trk == nil {
			panic(fmt.Sprintf("Complete() called on non-existent Task: %q", name))
		}

		// if the parent task is not nil, we need to increment the total count of the parent task
		if parentTrk != nil {
			parentTrk.counter.AddCompleted(1)
		}
		delete(trackers, prefix)

		// check to make sure that all children are complete
		if !trk.counter.Done() {
			panic(fmt.Sprintf("Attempting to Complete() %q but it sill has children", name))
		}

		dt := e.Time().Sub(trk.created)
		// if this is the root task, we don't need to return anything
		if prefix == "" {
			return ""
		}
		return fmt.Sprintf("%s ↦ Completed %s\n", strings.Join(name, separator), trk.FormatCompleted(dt))

	case *progressUpdate:
		log.Info("Processing progressUpdate") //nolint:sloglint
		// update progress bar data by adding the relative update
		if trk == nil {
			panic(fmt.Sprintf("Update() called on non-existent Task: %q", name))
		}

		if trk.tracker == nil {
			trk.tracker = tracker.NewByteTrackerFilter()
		}
		trk.tracker.Add(e.Time(), e.complete, e.total)

	default:
		panic("Unknown event type")
	}
	return ""
}

type taskTracker struct {
	name []string
	// TODO we could add the presentation form of the name to this struct
	// nameStr string // from strings.Join(name, separator)

	created time.Time
	tracker *tracker.ByteTrackerFilter
	counter tracker.Counter
}

func (tt *taskTracker) FormatCompleted(dt time.Duration) string {
	parts := make([]string, 0, 3)
	total := tt.counter.Total()
	if total > 0 {
		parts = append(parts, fmt.Sprintf("[%d]", total))
	}
	if tt.tracker != nil {
		// if no bytes were tracked, we can assume we used cache. Print "(cached)" for user
		if tt.tracker.Completed() == 0 {
			parts = append(parts, "(cached)")
		} else {
			parts = append(parts, tt.tracker.FormatCompleted(dt))
		}
	} else {
		parts = append(parts, fmt.Sprintf("in %v", dt.Round(time.Millisecond)))
	}
	return strings.Join(parts, " ")
}

func orderedTrackers(trackers map[string]*taskTracker) []taskTracker {
	// order of these alphabetically
	ordered := make([]taskTracker, 0, len(trackers))
	for _, trk := range trackers {
		// don't append the root tracker
		if len(trk.name) == 0 {
			continue
		}
		ordered = append(ordered, *trk)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		// TODO sort by length of name ([]string), then tie break on alphabetically by name
		a := ordered[i]
		b := ordered[j]
		if len(a.name) < len(b.name) {
			return true
		}
		if len(a.name) > len(b.name) {
			return false
		}
		// they are the same depth
		n := len(a.name) - 1
		return a.name[n] < b.name[n]
	})
	return ordered
}
