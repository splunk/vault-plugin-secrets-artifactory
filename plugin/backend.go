package artifactorysecrets

import (
	"context"

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
		Paths: framework.PathAppend(
			pathConfig(backend),
			pathRole(backend),
			pathRoleList(backend),
			pathToken(backend),
		),
	}

	// backend.artifactoryClient = newArtifactoryClient

	return backend
}

// TODO allow for mocked client
/*
func newArtifactoryClient(config *ConfigStorageEntry) (ArtifactoryClient, error) {

	c := &http.Client{} //nolint:ineffassign,staticcheck
	if config.BearerToken != "" {
		tp := transport.AccessTokenAuth{
			AccessToken: config.BearerToken,
		}
		c = tp.Client()
	} else if config.ApiKey != "" {
		tp := transport.ApiKeyAuth{
			ApiKey: config.ApiKey,
		}
		c = tp.Client()
	} else if config.Username != "" && config.Password != "" {
		tp := transport.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}
		c = tp.Client()
	} else {
		return ac, fmt.Errorf("bearer token, apikey or a pair of username/password isn't configured")
	}

	client, err := artifactory.NewClient(config.BaseURL, c)
	if err != nil {
		return ac, err
	}

	ac.client = client

	return ac, nil

}

*/
