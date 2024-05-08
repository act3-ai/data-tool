package pypi

import (
	"fmt"
	"net/http"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"git.act3-ace.com/ace/data/tool/internal/actions"
)

// Action represents a general pypi action.
type Action struct {
	*actions.DataTool

	// AllowYanked allows yanked packages to be mirrored
	AllowYanked bool
}

// PyPIClient create a http Client that pulls basic auth credentials from docker.
func (action *Action) PyPIClient() (remote.Client, error) {
	// TODO use the ace-dt config to find the docker config file
	// TODO should we implement the auth functionality in the http Transport/RoundTripper or the http Client?

	// TODO we could consider using https://github.com/zalando/go-keyring as the credential store (`act3-pt` uses this).

	// create the credential store
	storeOpts := credentials.StoreOptions{}
	store, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential store: %w", err)
	}

	// client := basicAuthClient{store, retry.DefaultClient}
	// HACK The auth.Client from ORAS hands basic auth like a regular client.
	// PyPI does not use anything else so the below approach works.
	client := auth.Client{
		Client: retry.DefaultClient,
		Header: http.Header{
			"User-Agent": {action.Config.UserAgent()},
		},
		Cache:      auth.NewSingleContextCache(),
		Credential: credentials.Credential(store),
	}

	return &client, nil
}

/*
type basicAuthClient struct {
	store  credentials.Store
	client remote.Client

	// TODO add an auth cache to avoid hitting the credential store too much
	// It can exec out to docker-credential-* so it can be expensive
	// cache  auth.Cache
}

func (c *basicAuthClient) Do(req *http.Request) (*http.Response, error) {
	// TODO Should we clone the request before modifying it?  req = req.Clone(req.Context())

	if u := req.URL.User; u != nil {
		// URL has credentials in it
		// These credentials seem to only be added on GET but not POST.
		// So we add then always here.
		if password, ok := u.Password(); ok {
			req.SetBasicAuth(u.Username(), password)
		}
		return c.client.Do(req) //nolint:wrapcheck
	}

	// Try the credential store
	cred, err := c.store.Get(req.Context(), req.URL.Host)
	if err != nil {
		return nil, fmt.Errorf("retreiving credentials for %q: %w", req.URL.Host, err)
	}
	if cred.Username != "" || cred.Password != "" {
		req.SetBasicAuth(cred.Username, cred.Password)
	}

	return c.client.Do(req) //nolint:wrapcheck
}
*/
