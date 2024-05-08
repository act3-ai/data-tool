// Package tracker provides progress tracking
package tracker

import (
	"fmt"
	"math"
	"time"

	"gitlab.com/act3-ai/asce/data/tool/internal/print"
)

// ByteTrackerFilter takes updates of the number of bytes processed/transferred and produces estimates of the time to complete and speed.
type ByteTrackerFilter struct {
	total    int64     // total bytes
	complete int64     // bytes completed
	t        time.Time // time of the most recent data

	filter alphaBetaFilter
}

// NewByteTrackerFilter constructs a new byte tracker filter.
func NewByteTrackerFilter() *ByteTrackerFilter {
	now := time.Now()
	return &ByteTrackerFilter{
		t:      now,
		filter: alphaBetaFilter{0.5, 0.1, 0, 0, now},
	}
}

// String implements fmt.Stringer interface.
func (bt *ByteTrackerFilter) String() string {
	return bt.Format(false)
}

// Format the byte tracker data.
func (bt *ByteTrackerFilter) Format(short bool) string {
	bt.filter.update(bt.t, float64(bt.complete))

	speedHumanized := print.Bytes(int64(math.Max(0, bt.filter.ẋ)))

	// calculate percentage
	var percentage float64
	var etc time.Duration
	if bt.total > 0 {
		// calculate ETC
		etc = bt.getETC()
		percentage = float64(bt.complete) / float64(bt.total) * 100
	}

	totalHumanized := print.Bytes(bt.total)
	completedHumanized := print.Bytes(bt.complete)

	var str string
	if short {
		// format progress string
		str = fmt.Sprintf("(%4.0f%%)", percentage)
	} else {
		// format progress string
		str = fmt.Sprintf("%s of %s (%4.1f%%) %s/s", completedHumanized, totalHumanized, percentage, speedHumanized)
	}

	if etc > time.Second {
		str = fmt.Sprintf("%s, ETC %v", str, etc)
	}

	return str
}

// FormatCompleted returns the final formatted completion message.
func (bt *ByteTrackerFilter) FormatCompleted(dt time.Duration) string {
	avgSpeed := float64(bt.complete) / float64(dt.Milliseconds())
	return fmt.Sprintf("%s in %v (%s/s)",
		print.Bytes(bt.complete),
		dt.Round(time.Millisecond),
		print.Bytes(int64(math.Max(0, avgSpeed))),
	)
}

// Total returns the total number of bytes.
func (bt *ByteTrackerFilter) Total() int64 {
	return bt.total
}

// Completed returns the number of completed bytes.
func (bt *ByteTrackerFilter) Completed() int64 {
	return bt.complete
}

// Add adds to the complete and total at the given time.
func (bt *ByteTrackerFilter) Add(t time.Time, complete, total int64) {
	bt.total += total
	bt.complete += complete
	if bt.t.Before(t) {
		bt.t = t
	}
}

func (bt *ByteTrackerFilter) getETC() time.Duration {
	var estimate float64

	if bt.total > bt.complete {
		estimate = float64(bt.total-bt.complete) / bt.filter.ẋ
	} // else the estimate is 0

	return time.Duration(estimate) * time.Second
}

// see https://en.wikipedia.org/wiki/Alpha_beta_filter
type alphaBetaFilter struct {
	ɑ, β float64   // tunable constants
	x, ẋ float64   // state at time t
	t    time.Time // time of the state
}

func (f *alphaBetaFilter) update(t time.Time, x float64) (float64, float64) {
	Δt := t.Sub(f.t).Seconds() //nolint:revive

	epsilon := math.Nextafter(1.0, 2.0) - 1.0
	if Δt <= epsilon {
		// panic("retrodiction is not supported")
		// skip the update
		return f.x, f.ẋ
	}

	// predict
	f.x += f.ẋ * Δt

	// residual
	r := x - f.x

	// update
	f.x += f.ɑ * r
	f.ẋ += f.β / Δt * r
	f.t = t

	return f.x, f.ẋ
}
