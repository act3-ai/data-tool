package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gitlab.com/act3-ai/asce/data/tool/internal/ui/tracker"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// debugUI is a UI used to record timestamped events for debugging.
type debugUI struct {
	// unbuffered channel, we want a crazy responsive UI, who cares about performance :)
	updates chan event

	// out is the file handle to write the debug output to
	out *os.File
}

// NewDebugUI returns a debug UI. Output is expected to be a log file.
func NewDebugUI(out *os.File) UI {
	return &debugUI{
		updates: make(chan event, bufferSize),
		out:     out,
	}
}

// debugStruct is used to keep track of a taskTracker and the log/csv files associated with the task.
type debugStruct struct {
	taskTracker *taskTracker
	rootDir     string
	logFile     *os.File
	counterCSV  *os.File
	progressCSV *os.File
}

// close will close the log file and the csv files.
func (d *debugStruct) close() error {
	if d.logFile != nil {
		if err := d.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}
	if d.counterCSV != nil {
		if err := d.counterCSV.Close(); err != nil {
			return fmt.Errorf("failed to close counter csv file: %w", err)
		}
	}
	if d.progressCSV != nil {
		if err := d.progressCSV.Close(); err != nil {
			return fmt.Errorf("failed to close progress csv file: %w", err)
		}
	}

	return nil
}

// trackerTotal returns the taskTracker's total.
func (d *debugStruct) trackerTotal() int64 {
	return d.taskTracker.tracker.Total()
}

// trackerCompleted returns the taskTracker's completed.
func (d *debugStruct) trackerCompleted() int64 {
	return d.taskTracker.tracker.Completed()
}

// counterAddTotal will add to the taskTracker's counter total, then write a counter csv update.
func (d *debugStruct) counterAddTotal(timestamp time.Duration) error {
	d.taskTracker.counter.AddTotal(1)
	return d.counterCSVUpdate(timestamp)
}

// counterAddCompleted will add to the taskTracker's counter completed, then write a counter csv update.
func (d *debugStruct) counterAddCompleted(timestamp time.Duration) error {
	d.taskTracker.counter.AddCompleted(1)
	return d.counterCSVUpdate(timestamp)
}

// counterTotal returns the counter's total.
func (d *debugStruct) counterTotal() int {
	return d.taskTracker.counter.Total()
}

// counterCompleted returns the counter's completed.
func (d *debugStruct) counterCompleted() int {
	return d.taskTracker.counter.Completed()
}

// CSVType is an enumeration of the supported CSV file types.
type CSVType int

const (
	// ProgressCSV suggests an *os.File is a csv for progress updates.
	ProgressCSV CSVType = iota
	// CounterCSV suggests an *os.File is a csv for task counting updates.
	CounterCSV
)

// addCSVFile will create the task's CSV file (progress or counter) and write the appropriate header.
func (d *debugStruct) addCSVFile(typeCSV CSVType) error {
	var fileNameCSV string
	// double pointer so we can assign to the pointer in the switch statement
	var fileCSV **os.File

	switch typeCSV {
	case ProgressCSV:
		fileNameCSV = "progress.csv"
		// same as d.progressCSV = *os.File
		fileCSV = &d.progressCSV
	case CounterCSV:
		fileNameCSV = "counter.csv"
		fileCSV = &d.counterCSV
	default:
		panic(fmt.Errorf("unsupported CSV type: %v", typeCSV))
	}

	// create task's CSV file
	file, err := os.Create(path.Join(d.rootDir, fileNameCSV))
	if err != nil {
		panic(fmt.Errorf("failed to create %s CSV file %s, err: %w", fileNameCSV, filepath.Join(d.rootDir, fileNameCSV), err))
	}
	*fileCSV = file

	// write header
	if _, err := (*fileCSV).WriteString("time,completed,total\n"); err != nil {
		panic(fmt.Errorf("failed to write header to %s CSV file %s, err: %w", fileNameCSV, filepath.Join(d.rootDir, fileNameCSV), err))
	}
	return nil
}

// addLogFile will create the task's log file and write the appropriate header.
func (d *debugStruct) addLogFile() error {
	// create task's log file
	logFile, err := os.Create(path.Join(d.rootDir, "log.jsonl"))
	if err != nil {
		return fmt.Errorf("failed to create log file %s, err: %w", filepath.Join(d.rootDir, "log.jsonl"), err)
	}
	d.logFile = logFile
	return nil
}

// progressCSVUpdate writes the current progress to the task's progress csv file.
func (d *debugStruct) progressCSVUpdate(timestamp time.Duration) error {
	if d.progressCSV != nil {
		// write to task's progress csv
		if _, err := fmt.Fprintf(d.progressCSV, "%v,%d,%d\n", int64(timestamp/time.Millisecond), d.trackerCompleted(), d.trackerTotal()); err != nil {
			return fmt.Errorf("failed to write progress to progress csv file %s, err: %w", filepath.Join(d.rootDir, "progress.csv"), err)
		}
	}
	return nil
}

// counterCSVUpdate writes the current count to the task's counter csv file.
func (d *debugStruct) counterCSVUpdate(timestamp time.Duration) error {
	if d.counterCSV == nil {
		return d.addCSVFile(CounterCSV)
	}
	// write to task's count csv
	if _, err := fmt.Fprintf(d.counterCSV, "%v,%d,%d\n", int64(timestamp/time.Millisecond), d.counterCompleted(), d.counterTotal()); err != nil {
		return fmt.Errorf("failed to write progress to counter csv file %s, err: %w", filepath.Join(d.rootDir, "counter.csv"), err)
	}
	return nil
}

// rootDirFromName takes a debug folder and a task name, and returns the root directory for the task (with some cleaning).
func rootDirFromName(debugFolder string, name []string) (string, error) {
	printName := strings.Join(name, separator)

	// Replace invalid characters with a safe character (e.g., '-')
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := reg.ReplaceAllString(printName, "-")

	// Trim leading and trailing spaces
	sanitized = strings.TrimSpace(sanitized)

	// replace spaces with underscores
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Check for maximum allowed length (e.g., 255 for most systems)
	if len(sanitized) > 255 {
		return "", errors.New("directory name is too long")
	}

	// Check if the resulting string is empty
	if len(sanitized) == 0 {
		return "", errors.New("directory name is empty after sanitization")
	}

	return filepath.Join(debugFolder, sanitized), nil
}

// newDebugStruct creates a new debugStruct given a taskUpdate and a root directory.
func newDebugStruct(update *taskUpdate, debugFolder string) *debugStruct {
	name := update.Name()
	if len(name) == 0 {
		name = []string{"ROOT_TASK"}
	}
	rootDir, err := rootDirFromName(debugFolder, name)
	if err != nil {
		panic(fmt.Errorf("failed to create debug folder given path %s, err: %w", rootDir, err))
	}

	// create directory for task
	if err := os.MkdirAll(rootDir, 0o777); err != nil {
		panic(fmt.Errorf("failed to create debug folder given path %s, err: %w", rootDir, err))
	}

	d := &debugStruct{
		taskTracker: &taskTracker{
			name:    name,
			created: update.Time(),
			tracker: nil,
			counter: *tracker.NewCounter(),
		},
		rootDir: rootDir,
	}

	// create task's log file
	if err := d.addLogFile(); err != nil {
		panic(err)
	}

	return d
}

const (
	startingTaskMessage   = "%v: Starting task: %s.\n"
	completedTaskMessage  = "%v: Completed task: %s.\n"
	addingProgressMessage = "%v: Adding progress to task: %s.\n"
	systemUpdateMessage   = "%v: SYSTEM update. NumGoroutine: %d\n"
)

// processInfoUpdate handles an infoUpdate event.
func (u *debugUI) processInfoUpdate(debugHelper *debugStruct, event *infoUpdate, printName string, timestamp time.Duration) error {
	if debugHelper == nil {
		return fmt.Errorf("Info() called on non-existent Task %q with message %s", printName, event.message)
	}
	// print message in json format with update timestamp to task's log file
	if _, err := fmt.Fprintf(debugHelper.logFile, `{"type": "%T", "name": "%s", "message": "%s", "timestamp": "%v"}%s`, *event, printName, event.message, timestamp, "\n"); err != nil {
		return err
	}
	return nil
}

// processTaskUpdate handles a taskUpdate event.
func (u *debugUI) processTaskUpdate(debugMap map[string]*debugStruct, debugHelper *debugStruct, debugParent *debugStruct, event *taskUpdate, debugFolder string, name []string, printName string, prefix string, timestamp time.Duration) error {
	if !event.done {
		// write to main log file that a new task was created
		if _, err := fmt.Fprintf(u.out, startingTaskMessage, timestamp, printName); err != nil {
			return err
		}

		// not done means we have a new Task
		// we increase the total count of the parent task if one exists
		if debugHelper != nil {
			return fmt.Errorf("non-unique task name provided: %q", name)
		}

		// create filename for csv file by cleaning the printName of all non-alphanumeric characters
		// create directory name for debug output
		debugHelper = newDebugStruct(event, debugFolder)
		debugMap[prefix] = debugHelper

		// if the parent task is not nil, we need to increment the total count of the parent task
		if debugParent != nil {
			if err := debugParent.counterAddTotal(timestamp); err != nil {
				return err
			}
		}
		// write task start to file
		if _, err := fmt.Fprintf(debugHelper.logFile, `{"type": "%T", "name": "%s", "message": "Starting task", "timestamp": "%v"}%s`, *event, printName, timestamp, "\n"); err != nil {
			return err
		}
	} else {
		// else; the task completed
		if debugHelper == nil {
			panic(fmt.Sprintf("Complete() called on non-existent Task: %q", name))
		}

		// write to main log file that a new task was created
		if _, err := fmt.Fprintf(u.out, completedTaskMessage, timestamp, printName); err != nil {
			return err
		}

		// if the parent task is not nil, we need to increment the total count of the parent task
		if debugParent != nil {
			if err := debugParent.counterAddCompleted(timestamp); err != nil {
				return err
			}
		}

		delete(debugMap, prefix)

		// check to make sure that all children are complete
		if !debugHelper.taskTracker.counter.Done() {
			panic(fmt.Sprintf("Attempting to Complete() %q but it sill has children", name))
		}

		dt := event.Time().Sub(debugHelper.taskTracker.created)
		message := "Completed " + debugHelper.taskTracker.FormatCompleted(dt)

		// print message in json format with update timestamp
		if _, err := fmt.Fprintf(debugHelper.logFile, `{"type": "%T", "name": "%s", "message": "%s", "timestamp": "%v"}%s`, *event, printName, message, timestamp, "\n"); err != nil {
			return err
		}

		// close the debugHelper, closing the log files and any created csv files
		if err := debugHelper.close(); err != nil {
			return err
		}
	}
	return nil
}

// processProgressUpdate handles a progressUpdate event.
func (u *debugUI) processProgressUpdate(debugHelper *debugStruct, event *progressUpdate, printName string, timestamp time.Duration) error {
	// update progress bar data by adding the relative update
	if debugHelper == nil {
		return fmt.Errorf("Update() called on non-existent Task: %q", printName)
	}

	// open csv file for this task if it is the first progress update
	if debugHelper.taskTracker.tracker == nil {
		// write to main log file that a new task was created
		if _, err := fmt.Fprintf(u.out, addingProgressMessage, timestamp, printName); err != nil {
			return err
		}

		debugHelper.taskTracker.tracker = tracker.NewByteTrackerFilter()
		// add progress csv to task
		if err := debugHelper.addCSVFile(ProgressCSV); err != nil {
			return err
		}
	}
	debugHelper.taskTracker.tracker.Add(event.Time(), event.complete, event.total)
	return debugHelper.progressCSVUpdate(timestamp)
}

// Run implements UI.
func (u *debugUI) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithGroup("UI")

	// get root debug folder from output path
	debugFolder := path.Dir(u.out.Name())

	debugMap := make(map[string]*debugStruct)
	t := time.NewTicker(time.Millisecond * 1000)
	startTime := time.Now()

	for {
		select {
		case update, ok := <-u.updates:
			if !ok {
				return nil
			}

			// process update common
			name := update.Name()
			printName := strings.Join(name, separator)
			prefix := strings.Join(name, null)
			timestamp := update.Time().Sub(startTime).Round(time.Millisecond)
			debugHelper := debugMap[prefix]
			var debugParent *debugStruct
			if len(name) > 0 {
				parent := strings.Join(name[:len(name)-1], null)
				debugParent = debugMap[parent]
			}

			switch event := update.(type) {
			case *infoUpdate:
				if err := u.processInfoUpdate(debugHelper, event, printName, timestamp); err != nil {
					return err
				}
			case *taskUpdate:
				if err := u.processTaskUpdate(debugMap, debugHelper, debugParent, event, debugFolder, name, printName, prefix, timestamp); err != nil {
					return err
				}
			case *progressUpdate:
				if err := u.processProgressUpdate(debugHelper, event, printName, timestamp); err != nil {
					return err
				}
			default:
				panic("Unknown event type")
			}
		case <-ctx.Done():
			logger.V(log, 2).InfoContext(ctx, "Context done")
			return ctx.Err()
		case <-t.C:
			// poll metrics we want for csv output
			// note that this does not guarantee that the system was polled at a regular interval
			// TODO: system metrics
			msgTime := time.Since(startTime).Round(time.Millisecond)
			if _, err := fmt.Fprintf(u.out, systemUpdateMessage, msgTime, runtime.NumGoroutine()); err != nil {
				return err
			}
		}
	}
}

// Shutdown implements UI.
func (u *debugUI) Shutdown() {
	close(u.updates)
}

// Root implements UI.
func (u *debugUI) Root(ctx context.Context) *Task {
	return newRootTask(logger.V(logger.FromContext(ctx).WithGroup("UI"), 1), u.updates)
}
