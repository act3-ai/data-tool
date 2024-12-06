package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
)

func Test_resolveSecret(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	var pass = "MyC001P4SSW0RD"

	t.Run("Plaintext", func(t *testing.T) {
		got, err := resolveSecret(ctx, pass)
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if got != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		key := "TEST_PASSWORD"
		t.Setenv(key, pass)

		got, err := resolveSecret(ctx, fmt.Sprintf("env:%s", key))
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if got != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})

	t.Run("File", func(t *testing.T) {
		passFile := filepath.Join(t.TempDir(), "testpass.txt")
		passFile, err := filepath.Abs(passFile)
		if err != nil {
			t.Errorf("resolving absolute password file path, error = %s", err)
		}
		if err := os.WriteFile(passFile, []byte(pass), 0666); err != nil {
			t.Errorf("initializing password file, error = %s", err)
			return
		}

		got, err := resolveSecret(ctx, fmt.Sprintf("file:%s", passFile))
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if got != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})

	t.Run("Command", func(t *testing.T) {
		got, err := resolveSecret(ctx, fmt.Sprintf("cmd:echo -n %s", pass))
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if got != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})
}
