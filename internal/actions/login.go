package actions

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
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

	Username string // Username credential for login
	Password string // Password credential for login
}

// Run runs the login action.
func (action *Login) Run(ctx context.Context, registry string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "login command activated")

	cfg := action.Config.Get(ctx)

	var err error
	if action.Username == "" || action.Password == "" {
		log.InfoContext(ctx, "Prompting for auth to registry", "reg", registry)
		if action.Username == "" {
			_, err = fmt.Fprint(out, "Username: ")
			if err != nil {
				return err
			}
			reader := bufio.NewReader(os.Stdin)
			line, _, err := reader.ReadLine()
			if err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
			action.Username = strings.TrimSpace(string(line))
		}

		if action.Password == "" {
			_, err = fmt.Fprint(out, "Password: ")
			if err != nil {
				return err
			}
			bpw, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("error reading password from term: %w", err)
			}
			action.Password = string(bpw)
			if action.Password == "" {
				return fmt.Errorf("password is required")
			}
		}
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
