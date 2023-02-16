// Copyright  2021 Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
		b.(*ArtifactoryBackend).client = &mockArtifactoryClient{}
	}

	return b, config.StorageView
}

// newArtAccEnv returns a new request and test backend with a real Artifactory configured
func newArtAccEnv(t *testing.T) (*logical.Request, logical.Backend) {
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

// newArtMockEnv returns a new request and test backend with mocked Artifactory client
func newArtMockEnv(t *testing.T) (*logical.Request, logical.Backend) {
	t.Helper()
	backend, storage := getTestBackend(t, true)

	req := &logical.Request{
		Storage: storage,
	}

	return req, backend

}
