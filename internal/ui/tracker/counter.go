package tracker

import "fmt"

// Counter represents a simple counter.
type Counter struct {
	total    int
	competed int
}

// NewCounter creates a new counter.
func NewCounter() *Counter {
	return &Counter{
		total:    0,
		competed: 0,
	}
}

// String implements fmt.Stringer interface.
func (c *Counter) String() string {
	return c.Format(false)
}

// Format the counter data.
func (c *Counter) Format(short bool) string {
	if c.total == 0 {
		return ""
	}

	if short {
		return fmt.Sprintf("[%d/%d]", c.competed, c.total)
	}
	return fmt.Sprintf("[%d of %d (%4.1f%%)]", c.competed, c.total, float64(c.competed)/float64(c.total)*100)
}

// Total returns the total number of tasks.
func (c *Counter) Total() int {
	return c.total
}

// Completed returns the number of completed tasks.
func (c *Counter) Completed() int {
	return c.competed
}

// AddTotal adds to the total count.
func (c *Counter) AddTotal(x int) {
	c.total += x
}

// AddCompleted adds to the completed count.
func (c *Counter) AddCompleted(x int) {
	c.competed += x
}

// Done returns true when all the tasks have completed.
func (c *Counter) Done() bool {
	return c.competed == c.total
}
