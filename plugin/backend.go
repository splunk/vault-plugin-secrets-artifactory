package artifactorysecrets

import (
	"context"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// ArtifactoryBackend is the backend for artifactory plugin
type ArtifactoryBackend struct {
	*framework.Backend
	view logical.Storage

	getClient func(ctx context.Context, config *ConfigStorageEntry) (Client, error)

	// Locks for guarding service clients
	// clientMutex sync.RWMutex

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
		getClient: NewClient,
	}

	backend.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        strings.TrimSpace(backendHelp),
		Paths: framework.PathAppend(
			pathConfig(backend),
			pathRole(backend),
			pathRoleList(backend),
			pathToken(backend),
		),
	}

	return backend
}

const backendHelp = `
The Artifactory secrets engine dynamically generates Artifactory access token
based on user defined permission targets. This enables users to gain access to
Artifactory resources without needing to create or manage a dedicated Artifactory
service account.

After mounting this secrets engine, you can configure the credentials using the
"config/" endpoints. You can generate roles using the "roles/" endpoints. You can 
then generate credentials for roles using the "token/" endpoints. 
`
