package ui

import (
	"context"
	"os"
)

func ExampleNewSimpleUI() {
	var ui = NewSimpleUI(os.Stdout)

	rootUI := ui.Root(context.Background())
	defer rootUI.Complete()

	gX := rootUI.SubTask("Processing X")
	defer gX.Complete()
	gX.Infof("Doing something now")

	gY := rootUI.SubTask("Processing Y")
	defer gY.Complete()
	gY.Infof("Doing other work")

	// some time has elapsed...

	p := gX.SubTaskWithProgress("Transferring")
	p.Update(4, 100)
	p.Infof("Finished X")
	p.Update(99, 100)
	p.Complete()
}

// Example UI outputs

/* Simple
Processing X
Doing something now
Transferring
ProgressBarThing
*/

/* Indent form
Processing X
	Doing something now
	Transferring
		ProgressBarThing
*/

/* Compact form
Processing X: Doing something now
Processing Y: Doing other work

some time later...
Processing X ==> Transferring ==> Progress: BarThing
*/

// Of course multi-paned TUIs are also possible
