package actions

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Login represents the login action.
type Login struct {
	*DataTool

	DisableAuthCheck bool // Disables the registry Ping() to check for auth.

	Username  string // Username credential for login
	Password  string // Password credential for login
	PassStdin bool
}

// Run runs the login action.
func (action *Login) Run(ctx context.Context, registry string, out io.Writer) error {
	log := logger.FromContext(ctx).With("reg", registry)
	log.InfoContext(ctx, "login command activated")

	cfg := action.Config.Get(ctx)

	if action.PassStdin {
		pass, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading password from stdin: %w", err)
		}
		action.Password = strings.TrimSuffix(string(pass), "\n")
		action.Password = strings.TrimSuffix(action.Password, "\r")
	}

	var err error
	if action.Username == "" {
		action.Username, err = promptUsername(ctx, out)
		if err != nil {
			return fmt.Errorf("prompting username input: %w", err)
		}
	}

	if action.Password == "" {
		action.Password, err = promptPassword(ctx, out)
		if err != nil {
			return fmt.Errorf("prompting password input: %w", err)
		}
	}

	action.Password, err = resolveSecret(ctx, action.Password)
	if err != nil {
		return fmt.Errorf("resolving secret: %w", err)
	}

	store, err := credentials.NewStore(cfg.RegistryAuthFile, credentials.StoreOptions{
		AllowPlaintextPut: true,
	})
	if err != nil {
		return fmt.Errorf("opening credential store at %q: %w", cfg.RegistryAuthFile, err)
	}

	// Get the remote.Registry
	reg, err := action.Config.NewRegistry(ctx, registry)
	// reg, err := remote.NewRegistry(registry)
	if err != nil {
		return fmt.Errorf("parsing registry %q: %w", registry, err)
	}

	cred := auth.Credential{
		Username: action.Username,
		Password: action.Password,
	}

	if action.DisableAuthCheck {
		hostname := credentials.ServerAddressFromRegistry(registry)
		if err := store.Put(ctx, hostname, cred); err != nil {
			return fmt.Errorf("failed to store the credentials for %s: %w", hostname, err)
		}

		_, err = fmt.Fprintln(out, "Credentials stored")
		if err != nil {
			return err
		}
	} else {
		err = credentials.Login(ctx, store, reg, cred)
		if err != nil {
			return fmt.Errorf("logging in: %w", err)
		}

		_, err = fmt.Fprintln(out, "Login successful")
		if err != nil {
			return err
		}
	}

	log.InfoContext(ctx, "login command completed")
	return nil
}

func promptUsername(ctx context.Context, out io.Writer) (string, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "prompting username for registry auth")
	_, err := fmt.Fprint(out, "Username: ")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}
	username := strings.TrimSpace(string(line))

	return username, nil
}

func promptPassword(ctx context.Context, out io.Writer) (string, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "prompting password for registry auth")
	_, err := fmt.Fprint(out, "Password: ")
	if err != nil {
		return "", err
	}
	bpw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("reading password from term: %w", err)
	}
	password := string(bpw)
	if password == "" {
		return "", fmt.Errorf("password is required")
	}

	return password, nil
}

const (
	envSecretSource     = "env"  // env:PASSWORD where $PASSWORD=MyC001P4ssw0rd
	fileSecretSource    = "file" // file:/home/user/password.txt ; an absolute path
	commandSecretSource = "cmd"  // cmd:echo -n MyC001P4ssw0rd
)

// resolveSecret checks for special cases where a secret was not passed directly,
// e.g. via an env var 'env:PASSWORD' whose value is the secret itself. If there
// is no special case the original value is returned, which is considered the
// secret itself.
func resolveSecret(ctx context.Context, value string) (string, error) {
	// modified version of https://github.com/dagger/dagger/blob/main/cmd/dagger/flags.go#L505
	log := logger.FromContext(ctx)
	var plaintext string

	prefix, val, ok := strings.Cut(value, ":")
	if !ok {
		log.InfoContext(ctx, "secret provided directly")
		return prefix, nil
	}

	switch prefix {
	case envSecretSource:
		log.InfoContext(ctx, "reading secret from environment variable")
		envPlaintext, ok := os.LookupEnv(val)
		if !ok {
			// Don't show the entire env var name, in case the user accidentally passed the value instead...
			// This is important because users originally *did* have to pass the value, before we changed to
			// passing by name instead.
			key := val
			if len(key) >= 4 {
				key = key[:3] + "..."
			}
			return "", fmt.Errorf("secret env var not found: %q", key)
		}
		plaintext = envPlaintext

	case fileSecretSource:
		log.InfoContext(ctx, "reading secret from file")
		filePlaintext, err := os.ReadFile(val)
		if err != nil {
			return "", fmt.Errorf("failed to read secret file %q: %w", val, err)
		}
		plaintext = string(filePlaintext)

	case commandSecretSource:
		var stdoutBytes []byte
		var err error
		if runtime.GOOS == "windows" { // TODO: Test on windows, we're trusting dagger here...
			stdoutBytes, err = exec.CommandContext(ctx, "cmd.exe", "/C", val).Output()
		} else {
			// #nosec G204
			stdoutBytes, err = exec.CommandContext(ctx, "sh", "-c", val).Output()
		}
		if err != nil {
			return "", fmt.Errorf("failed to run secret command %q: %w", val, err)
		}
		plaintext = string(stdoutBytes)

	default:
		return "", fmt.Errorf("unsupported secret arg source: %q", prefix)
	}

	return plaintext, nil
}
