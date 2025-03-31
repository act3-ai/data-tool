package secret

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	tlog "gitlab.com/act3-ai/asce/go-common/pkg/test"
)

func Test_resolveSecret(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	var pass = "MyC001P4SSW0RD"

	t.Run("Plaintext", func(t *testing.T) {
		v := &Value{}
		if err := v.Set(pass); err != nil {
			t.Errorf("setting secret, error = %v", err)
			return
		}

		_, err := v.resolveSecret(ctx)
		if err == nil {
			t.Errorf("resolveSecret() expected error, got nil error")
			return
		}
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		key := "TEST_PASSWORD"
		t.Setenv(key, pass)

		v := &Value{}
		if err := v.Set(fmt.Sprintf("env:%s", key)); err != nil {
			t.Errorf("setting secret, error = %v", err)
			return
		}

		got, err := v.resolveSecret(ctx)
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if string(got) != pass {
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

		v := &Value{}
		if err := v.Set(fmt.Sprintf("file:%s", passFile)); err != nil {
			t.Errorf("setting secret, error = %v", err)
			return
		}

		got, err := v.resolveSecret(ctx)
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if string(got) != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})

	t.Run("Command", func(t *testing.T) {
		v := &Value{}
		if err := v.Set(fmt.Sprintf("cmd:echo %s", pass)); err != nil {
			t.Errorf("setting secret, error = %v", err)
			return
		}

		got, err := v.resolveSecret(ctx)
		if err != nil {
			t.Errorf("resolveSecret() error = %v", err)
			return
		}
		if string(got) != pass {
			t.Errorf("resolveSecret() got = %s, want = %s", got, pass)
			return
		}
	})
}
