package artifactorysecrets

import (
	"context"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// ArtifactoryBackend is the backend for artifactory plugin
type ArtifactoryBackend struct {
	*framework.Backend
	view logical.Storage

	// Locks for guarding service clients
	clientMutex sync.RWMutex

	roleLocks []*locksutil.LockEntry
}

// Factory is factory for backend
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

// Backend export the function to create backend and configure
func Backend(conf *logical.BackendConfig) *ArtifactoryBackend {
	backend := &ArtifactoryBackend{
		view:      conf.StorageView,
		roleLocks: locksutil.CreateLocks(),
	}

	backend.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Paths: framework.PathAppend(
			pathConfig(backend),
			pathRole(backend),
			pathRoleList(backend),
			pathToken(backend),
		),
	}

	return backend
}
