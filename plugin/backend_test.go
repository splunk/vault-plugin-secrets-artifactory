package artifactorysecrets

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

// getTestBackend returns the mocked out backend for testing
func getTestBackend(t *testing.T, mockArtifactory bool) (logical.Backend, logical.Storage) {
	t.Helper()
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b, err := Factory(context.Background(), config)
	require.NoError(t, err, "unable to create backend")

	if mockArtifactory {
		b.(*ArtifactoryBackend).getClient = func(ctx context.Context, c *ConfigStorageEntry) (Client, error) {
			return &mockArtifactoryClient{}, nil
		}
	}

	return b, config.StorageView
}

// newAccEnv returns a new request and test backend with a real Artifactory configured
func newAccEnv(t *testing.T) (*logical.Request, logical.Backend) {
	t.Helper()

	backend, storage := getTestBackend(t, false)

	conf := map[string]interface{}{
		"base_url":     os.Getenv("ARTIFACTORY_URL"),
		"bearer_token": os.Getenv("ARTIFACTORY_BEARER_TOKEN"),
		"max_ttl":      "3600s",
	}

	testConfigUpdate(t, backend, storage, conf)

	req := &logical.Request{
		Storage: storage,
	}

	return req, backend
}

// newMockEnv returns a new request and test bacekdn with mocked Artifactory client
func newMockEnv(t *testing.T) (*logical.Request, logical.Backend) {
	t.Helper()
	backend, storage := getTestBackend(t, true)

	req := &logical.Request{
		Storage: storage,
	}

	return req, backend

}
