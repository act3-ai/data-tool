package actions

import (
	"context"
	"fmt"
	"io"

	"oras.land/oras-go/v2/registry/remote/credentials"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Logout represents the logout action.
type Logout struct {
	*DataTool

	Username string // Username credential for logout
	Password string // Password credential for logout
}

// Run runs the logout action.
func (action *Logout) Run(ctx context.Context, registry string, out io.Writer) error {
	log := logger.FromContext(ctx)

	cfg := action.Config.Get(ctx)

	log.InfoContext(ctx, "logout command activated")

	store, err := credentials.NewStore(cfg.RegistryAuthFile, credentials.StoreOptions{})
	if err != nil {
		panic(err)
	}
	err = credentials.Logout(ctx, store, registry)
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintln(out, "logout completed")
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "logout command completed")

	return nil
}
