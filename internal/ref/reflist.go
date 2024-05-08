package ref

import (
	"context"
	"errors"
)

// ErrRefListEmpty is an error returned when querying a list of registries, repositories, or refs that indicates that
// the resulting list was empty (Versus another kind of error).
var ErrRefListEmpty = errors.New("repository not found")

// RepositoryListProvider defines an interface for retrieving a string array list of repositories for a registry, as
// implemented by RegistryRepos. The registry info is provided as a Ref in order to also include the
// communication scheme if necessary.
type RepositoryListProvider interface {
	// RepositoryList retrieves a list of known repositories based on a ref containing a registry and transfer scheme
	RepositoryList(registry Ref) ([]string, error)
}

// ListProvider defines an interface for retrieving a full list of references based on a provided registry and
// list of repositories.  An empty repository list should be interpreted as "all" repositories.  The registry info
// is provided as a Ref in order to also include the communication scheme if necessary.
type ListProvider interface {
	RepositoryListProvider

	// RefList accepts a Ref defining a registry and a list of Repositories, and return a list of Ref that contains
	// full references for all matching repositories on the registry, including all known tags. If repoList is empty
	// a list of all known repositories is first obtained using the RepositoryListProvider.
	RefList(ctx context.Context, registry Ref, repoList []string) ([]Ref, error)
}
