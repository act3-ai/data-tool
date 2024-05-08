package ui

import (
	"time"
)

type event interface {
	// Time when the event occurred
	Time() time.Time

	// Name is the name of the task that this message belongs (the prefix)
	Name() []string
}

type eventBase struct {
	t    time.Time
	name []string
}

func (e eventBase) Time() time.Time {
	return e.t
}

func (e eventBase) Name() []string {
	return e.name
}

// eventBase implements the event interface.
var _ event = (*eventBase)(nil)

// infoUpdate is the data that represents an informational message event.
type infoUpdate struct {
	eventBase
	message string
}

// taskUpdate is the data that represents a task.
type taskUpdate struct {
	eventBase
	done bool
}

// progressUpdate is the data that represents a progress update.
type progressUpdate struct {
	eventBase
	complete, total int64
}
