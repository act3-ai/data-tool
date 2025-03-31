package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"

	"gitlab.com/act3-ai/asce/data/tool/internal/secret"
)

// Login represents the login action.
type Login struct {
	*DataTool

	DisableAuthCheck bool // Disables the registry Ping() to check for auth.

	Username  string               // Username credential for login
	Password  secret.ValueResolver // Password credential for login
	PassStdin bool
}

// Run runs the login action.
func (action *Login) Run(ctx context.Context, registry string, out io.Writer) error {
	// sanity
	if action.PassStdin && action.Password.String() != "" {
		return fmt.Errorf("passowrd may be provided (indirectly) by flag or through stdin, not both")
	}

	log := logger.FromContext(ctx).With("reg", registry)
	log.InfoContext(ctx, "login command activated")

	var err error
	if action.Username == "" {
		action.Username, err = secret.PromptUsername(ctx, out)
		if err != nil {
			return fmt.Errorf("prompting username input: %w", err)
		}
	}

	var pass redact.Secret
	switch {
	case action.PassStdin:
		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading password from stdin: %w", err)
		}
		p := strings.TrimSuffix(string(in), "\n")
		p = strings.TrimSuffix(p, "\r")
		pass = redact.Secret(p)
	case action.Password.String() == "":
		pass, err = secret.PromptPassword(ctx, out)
		if err != nil {
			return fmt.Errorf("prompting password input: %w", err)
		}
	default:
		// 'env:', 'file:', or 'cmd:'
		pass, err = action.Password.Get(ctx)
		if err != nil {
			return fmt.Errorf("resolving secret: %w", err)
		}
	}

	cfg := action.Config.Get(ctx)
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
		Password: string(pass),
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
