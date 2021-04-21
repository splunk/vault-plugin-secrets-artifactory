package artifactorysecrets

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

// return the mocked out backend for testing
func getTestBackend(t *testing.T) (logical.Backend, logical.Storage) {
	t.Helper()
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b := Backend(config)
	if err := b.Setup(context.Background(), config); err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	// cfg, err := b.getConfig(context.Background(), config.StorageView)
	// if err != nil {
	// 	t.Fatal("unable to create backend: %v", err)
	// }

	b.getClient = func(ctx context.Context, c *ConfigStorageEntry) (Client, error) {
		return &mockArtifactoryClient{}, nil
	}

	// b.getClient = func(ctx context.Context, storage logical.Storage) (Client, error) {
	// 	return &mockArtifactoryClient{}, nil
	// }

	return b, config.StorageView
}
