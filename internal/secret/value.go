// Package secret provides utility functions for handling secrets.
package secret

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
)

// Secret is the type of a secret value.
const Secret = "Secret"

// ValueResolver extends pflag.Value with a secret value getter.
type ValueResolver interface {
	pflag.Value

	// Get returns the secret value.
	Get(context.Context) (redact.Secret, error)
}

// Value implements ValueResolver.
type Value struct {
	// modified version of https://github.com/dagger/dagger/blob/main/cmd/dagger/flags.go#L470
	source    secretSource
	sourceVal string
	secret    redact.Secret
}

// Type returns the type of the pflag.Value, i.e. "Secret".
func (v *Value) Type() string {
	return Secret
}

// Set parses a secret source into the source prefix and value.
func (v *Value) Set(s string) error {
	src, val, ok := strings.Cut(s, ":")
	if !ok {
		// case of e.g. `MY_ENV_SECRET`, which is shorthand for `env:MY_ENV_SECRET`
		val = src
		src = string(envSrc)
	}
	v.source = secretSource(src)
	v.sourceVal = val

	return nil
}

// String returns the original secret source provided to Set().
func (v *Value) String() string {
	if v.sourceVal == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", v.source, v.sourceVal)
}

// Get returns the value of the secret resolved from the secret source.
func (v *Value) Get(ctx context.Context) (redact.Secret, error) {
	if v.secret != "" {
		return v.secret, nil
	}
	return v.resolveSecret(ctx)
}

type secretSource string

const (
	envSrc  secretSource = "env"  // env:PASSWORD; where $PASSWORD=MyC001P4ssw0rd
	fileSrc secretSource = "file" // file:/home/user/password.txt ; an absolute path
	cmdSrc  secretSource = "cmd"  // cmd:secret-tool lookup username exampleuser server reg.example.com
)

var errUnsupportedSecretSource = fmt.Errorf("unsupported secret source, want '%s', '%s', or '%s'", envSrc, fileSrc, cmdSrc)

// resolveSecret resolves a secret value based off of prefixes
// that identify the source of a secret, i.e. a secretsource.
func (v *Value) resolveSecret(ctx context.Context) (redact.Secret, error) {
	// modified version of https://github.com/dagger/dagger/blob/main/cmd/dagger/flags.go#L505
	log := logger.FromContext(ctx)
	var plaintext redact.Secret

	switch v.source {
	case envSrc:
		log.InfoContext(ctx, "reading secret from environment variable")
		envPlaintext, ok := os.LookupEnv(v.sourceVal)
		if !ok {
			// Don't show the entire env var name, in case the user accidentally passed the value instead
			key := v.sourceVal
			if len(key) >= 4 {
				key = key[:3] + "..."
			}
			return "", fmt.Errorf("secret env var not found: %q", key)
		}
		plaintext = redact.Secret(envPlaintext)

	case fileSrc:
		log.InfoContext(ctx, "reading secret from file")
		filePlaintext, err := os.ReadFile(v.sourceVal)
		if err != nil {
			return "", fmt.Errorf("failed to read secret file %q: %w", v.sourceVal, err)
		}
		plaintext = redact.Secret(filePlaintext)

	case cmdSrc:
		var stdoutBytes []byte
		var err error
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if runtime.GOOS == "windows" { // TODO: Test on windows, we're trusting dagger here...
			stdoutBytes, err = exec.CommandContext(ctx, "cmd.exe", "/C", v.sourceVal).Output()
		} else {
			// #nosec G204
			stdoutBytes, err = exec.CommandContext(ctx, "sh", "-c", v.sourceVal).Output()
		}
		if err != nil {
			return "", fmt.Errorf("failed to run secret command %q: %w", v.sourceVal, err)
		}
		plaintext = redact.Secret(bytes.TrimSpace(stdoutBytes))

	default:
		return "", fmt.Errorf("%w: got %q", errUnsupportedSecretSource, v.source)
	}

	return plaintext, nil
}
