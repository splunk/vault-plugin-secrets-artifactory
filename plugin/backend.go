package artifactorysecrets

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// ArtifactoryBackend is the backend for artifactory plugin
type ArtifactoryBackend struct {
	*framework.Backend
	view      logical.Storage
	client    Client
	lock      sync.RWMutex
	roleLocks []*locksutil.LockEntry
}

func (b *ArtifactoryBackend) getClient(ctx context.Context, s logical.Storage) (Client, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client.Valid() {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	if b.client.Valid() {
		return b.client, nil
	}

	config, err := b.getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	c, err := NewClient(ctx, config)
	if err != nil {
		return nil, err
	}
	b.client = c

	return c, nil
}

func (b *ArtifactoryBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.client = nil
}

func (b *ArtifactoryBackend) invalidate(ctx context.Context, key string) {
	switch key {
	case configPrefix:
		b.reset()
	}
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
		client:    NewArtifactoryClient(),
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
		Invalidate: backend.invalidate,
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
