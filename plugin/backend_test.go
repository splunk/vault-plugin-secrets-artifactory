package artifactorysecrets

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

// return the mocked out backend for testing
func getTestBackend(t *testing.T) (logical.Backend, logical.Storage) {
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b, err := Factory(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	return b, config.StorageView
}
