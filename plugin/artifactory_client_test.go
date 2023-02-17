// Copyright  2021 Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envVarRunArtAccTests = "ARTIFACTORY_ACC"

func TestNewClientFail(t *testing.T) {
	t.Parallel()
	t.Run("no_config", func(t *testing.T) {

		c, err := NewClient(nil)
		assert.Error(t, err, "nil config should thrown an error when retrieving Artifactory client")
		assert.Nil(t, c)
	})

	t.Run("empty_config", func(t *testing.T) {
		config := &ConfigStorageEntry{}
		c, err := NewClient(config)
		assert.Error(t, err, "NewClient should return an error if config is missing auth")
		assert.Nil(t, c, "NewClient should return nil client on error")
	})
}

func TestArtAccNewClient(t *testing.T) {
	t.Parallel()
	if os.Getenv(envVarRunArtAccTests) == "" {
		t.Skip("skipping Artifactory acceptance test (ARTIFACTORY_ACC env var not set)")
	}

	baseUrl := os.Getenv("ARTIFACTORY_URL")
	bearerToken := os.Getenv("ARTIFACTORY_BEARER_TOKEN")
	username := os.Getenv("ARTIFACTORY_USER")
	password := os.Getenv("ARTIFACTORY_PASSWORD")
	apiKey := os.Getenv("ARTIFACTORY_API_KEY")

	require := require.New(t)
	require.NotEmpty(baseUrl)
	require.NotEmpty(bearerToken)
	require.NotEmpty(username)
	require.NotEmpty(password)
	require.NotEmpty(apiKey)
	require.True(strings.HasSuffix(baseUrl, "/"))

	tests := []struct {
		name   string
		config *ConfigStorageEntry
	}{
		{
			name: "bearer_token",
			config: &ConfigStorageEntry{
				BaseURL:     baseUrl,
				BearerToken: bearerToken,
			},
		},
		{
			name: "user_pass",
			config: &ConfigStorageEntry{
				BaseURL:  baseUrl,
				Username: username,
				Password: password,
			},
		},
		{
			name: "api_key",
			config: &ConfigStorageEntry{
				BaseURL: baseUrl,
				ApiKey:  apiKey,
			},
		},
	}

	for _, test := range tests {
		test := test // capture range var
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewClient(test.config)
			assert.NoError(t, err)
			assert.NotNil(t, c)
			assert.True(t, c.Valid())
			ac, ok := c.(*artifactoryClient)
			require.True(ok)

			// call an api endpoint to verify working auth
			users, err := ac.client.GetAllUsers()
			require.NoError(err)
			require.NotNil(users)
		})
	}
}

func TestValid(t *testing.T) {

	tests := []struct {
		name     string
		client   *artifactoryClient
		asserter assert.BoolAssertionFunc
	}{
		{
			name: "valid",
			client: &artifactoryClient{
				expiration: time.Now().Add(clientTTL),
			},
			asserter: assert.True,
		},
		{
			name: "expired ttl",
			client: &artifactoryClient{
				expiration: time.Now().Add(-1 * time.Minute),
			},
			asserter: assert.False,
		},
	}

	for _, test := range tests {
		test := test // capture range var
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.asserter(t, test.client.Valid())
		})
	}
}

type mockArtifactoryClient struct{}

var _ Client = &mockArtifactoryClient{}

func (ac *mockArtifactoryClient) Valid() bool {
	return true
}

func (ac *mockArtifactoryClient) CreateOrReplaceGroup(role *RoleStorageEntry) error {
	return nil
}

func (ac *mockArtifactoryClient) DeleteGroup(role *RoleStorageEntry) error {
	return nil
}
func (ac *mockArtifactoryClient) CreateOrUpdatePermissionTarget(role *RoleStorageEntry, pt *PermissionTarget, ptName string) error {
	return nil
}
func (ac *mockArtifactoryClient) DeletePermissionTarget(ptName string) error {
	return nil
}
func (ac *mockArtifactoryClient) CreateToken(tokenReq TokenCreateEntry, role *RoleStorageEntry) (auth.CreateTokenResponseData, error) {
	return auth.CreateTokenResponseData{}, nil
}

// getAccClient returns the underlying artifactory services manager for full access to the Artifactory API.
// This is used in integration tests to validate permission targets and groups.
func mustGetAccClient(ctx context.Context, t *testing.T, req *logical.Request, b logical.Backend) artifactory.ArtifactoryServicesManager {
	t.Helper()

	backend, ok := b.(*ArtifactoryBackend)
	require.True(t, ok, "invalid backend implementation")

	ac, err := backend.getClient(ctx, req.Storage)
	require.NoError(t, err, "Artifactory client error: %s", err)

	// get the actual Jfrog Client
	acImpl, ok := ac.(*artifactoryClient)
	require.True(t, ok, "invalid artifactory client implementation")

	return acImpl.client
}
