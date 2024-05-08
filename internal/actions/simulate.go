package actions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"git.act3-ace.com/ace/data/tool/internal/ui"
)

// Simulate is an action for simulating UI activity.
// Used for testing the UI machinery.
type Simulate struct {
	*DataTool
	NumTasks          int
	NumMaxParallel    int
	NumCountRecursive int
}

// Run runs the visualize action.
func (action *Simulate) Run(ctx context.Context) error {
	rootUI := ui.FromContextOrNoop(ctx)
	rootUI.Infof("This is the root task")

	// create an error group for parallel task execution
	g, _ := errgroup.WithContext(ctx)

	g.SetLimit(action.NumMaxParallel)

	for i := 0; i < action.NumTasks; i++ {
		id := i + 1
		subU1 := rootUI.SubTask(fmt.Sprintf("Task %d", id))
		subU1.Infof("Starting task %d", id)
		g.Go(func() error {
			// fake work
			subU1.Infof("working...")
			time.Sleep(time.Duration(id*5) * time.Second)

			subU1.Complete()
			return nil
		})
	}

	for i := 0; i < 2; i++ {
		id := i + 1
		subU2 := rootUI.SubTaskWithProgress(fmt.Sprintf("Upload %d", id))
		subU2.Infof("Starting %d", id)

		var n, step int64 = 200000, 20
		g.Go(func() error {
			// fake work
			subU2.Update(0, n)
			for j := int64(0); j < n; j += step {
				time.Sleep(1 * time.Millisecond)
				subU2.Update(step, 0)
			}
			subU2.Complete()
			return nil
		})
	}

	for i := 0; i < 2; i++ {
		id := i + 1
		subU3 := rootUI.SubTaskWithProgress(fmt.Sprintf("Download %d", id))
		subU3.Infof("Starting %d", id)

		var n, step int64 = 400000, 20
		g.Go(func() error {
			// fake work
			subU3.Update(0, n)
			for j := int64(0); j < n; j += step {
				time.Sleep(1 * time.Millisecond)
				subU3.Update(step, 0)
			}
			subU3.Complete()
			return nil
		})
	}

	counterUI := rootUI.SubTask("Stuff")
	defer counterUI.Complete()
	makeTaskLoop(action.NumCountRecursive, 2, counterUI)

	for i := 0; i < action.NumTasks; i++ {
		subUI := counterUI.SubTask("task_" + strconv.Itoa(i+1))
		time.Sleep(1 * time.Second)
		g.Go(func() error {
			subUI.Infof("working...")
			defer subUI.Complete()
			return runCounterTasks(ctx, subUI, i+1)
		})
	}

	return g.Wait()
}

// makeTaskLoop will create a simple for loop from 0 to the loopTotal that will run fake work on the given task
// the task can create sub-tasks if the recursive int is > 0
// the task will complete when the loop is done
// if the goroutine flag is set, the loop will run in a go routine that blocks until the task is complete.
func makeTaskLoop(recursive int, loopTotal int, task *ui.Task) {
	for i := 0; i < loopTotal; i++ {
		if recursive > 0 {
			name := "recursive_" + strconv.Itoa(recursive) + "_loop_" + strconv.Itoa(i+1)
			subTask := task.SubTask(name)
			makeTaskLoop(recursive-1, loopTotal, subTask)
			subTask.Complete()
		} else {
			task.Infof("working on part %d", i+1)
			// time.Sleep(1 * time.Second)
		}
	}
}

func runCounterTasks(ctx context.Context, subUI *ui.Task, total int) error {
	// create an error group for parallel task execution
	g, _ := errgroup.WithContext(ctx)

	for j := 0; j < total; j++ {
		subUI2 := subUI.SubTask("Subtask_" + strconv.Itoa(j+1))
		g.Go(func() error {
			subUI2.Infof("working...")
			defer subUI2.Complete()
			for k := 0; k < 10; k++ {
				subUI2.Infof("part %d", k)
				time.Sleep(1 * time.Second)
			}
			return nil
		})
		time.Sleep(1 * time.Second)
	}
	return g.Wait()
}
