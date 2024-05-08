package bottle

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Edit represents the bottle edit action.
type Edit struct {
	*Action

	Write WriteBottleOptions
}

// Run runs the bottle edit action.
func (action *Edit) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Edit command activated")

	cfg := action.Config.Get(ctx)

	bottlePath := action.Dir
	log.InfoContext(ctx, "Using path", "path", bottlePath)

	if err := checkBottle(ctx, bottlePath, ""); err != nil {
		return err
	}

	// TODO call btl.GetPrettyYAML...() []byte instead
	filePath := bottle.EntryFile(bottlePath)

	// create temporary file and copy content into it
	editFile := filePath + ".temp"

	if err := util.CopyFile(filePath, editFile); err != nil {
		return fmt.Errorf("error copying data bottle configuration into temporary file: %w", err)
	}

	// Start editing loop
	keepEditing := false

	for ok := true; ok; ok = keepEditing {

		if err := OpenFileInEditor(ctx, cfg.Editor, editFile); err != nil {
			return fmt.Errorf("error loading temporary data bottle configuration for editing: %w", err)
		}

		validationErr := checkBottle(ctx, bottlePath, editFile)

		if validationErr == nil {
			keepEditing = false
			// Replace the config file with the updated copy (os.Rename automatically source file)
			log.InfoContext(ctx, "Edits made were valid")
			if err := os.Rename(editFile, filePath); err != nil {
				return fmt.Errorf("error renaming config: %w", err)
			}
			log.InfoContext(ctx, "Edits were saved to disk")

		} else if validationErr != nil {

			if _, err := fmt.Fprintln(out, validationErr.Error()); err != nil {
				return err
			}

			userAnswer := promptForMoreEdit(out)

			if userAnswer == preserveEdit {
				keepEditing = true
			} else if userAnswer == discardEdit {
				keepEditing = false
				if err := os.Remove(editFile); err != nil {
					return fmt.Errorf("failed to remove temporary file %s from bottle folder %s: %w", editFile, bottlePath, err)
				}
				log.InfoContext(ctx, "Edits made were discarded because they were invalid")
			}
		}
	}

	log.InfoContext(ctx, "Edit command completed")

	return nil
}

// OpenFileInEditor opens the specified file in the environment's
// default editor, or the config.DefaultEditor for user modification.
func OpenFileInEditor(ctx context.Context, editor string, filename string) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Opening editor", "editor", editor)

	// Get the full executable path for the editor.
	executable, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("error obtaining path to editor executable: %w", err)
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running executable: %w", err)
	}
	return nil
}

// constants used to track users decision about continuing editing or discarding changes.
const (
	preserveEdit = iota
	discardEdit
)

// promptForMoreEdit prompts users, and captures their decision on whether to continue or stop editing entry.yaml.
func promptForMoreEdit(out io.Writer) int {
	validAnswer := false
	userAnswer := discardEdit

	if _, err := fmt.Fprintln(out, " The edits made were invalid."); err != nil {
		return discardEdit
	}

	for ok := true; ok; ok = !validAnswer {

		if _, err := fmt.Fprint(out, " > Do you want to continue editing (c) or discard changes (d): "); err != nil {
			return discardEdit
		}

		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			return discardEdit
		}

		switch {
		case strings.EqualFold(answer, "c"):
			validAnswer = true
			userAnswer = preserveEdit
		case strings.EqualFold(answer, "d"):
			if _, err := fmt.Fprintln(out, " We're discarding your changes, and restoring the previous file."); err != nil {
				return discardEdit
			}
			validAnswer = true
			userAnswer = discardEdit
		default:
			if _, err := fmt.Fprintln(out, " You entered an invalid option. Please try again."); err != nil {
				return discardEdit
			}
		}
	}

	return userAnswer
}

// checkBottle checks for errors within a bottle configuration.  bottlePath specifies a bottle directory to validate
// and filePath is an optional path to a yaml file containing bottle configuration data.  If the filePath is empty, the
// default entry.yaml file within the bottlePath is used.
func checkBottle(ctx context.Context, bottlePath string, filePath string) error {
	btl, err := bottle.LoadBottle(bottlePath,
		bottle.DisableDestinationCreate(true),
		bottle.DisableCache(true),
	)
	if err != nil {
		return err
	}
	return btl.Definition.ValidateWithContext(ctx)
}
