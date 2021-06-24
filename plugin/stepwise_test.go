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
	"fmt"
	"os"
	"sync"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/testing/stepwise"
	"github.com/stretchr/testify/require"
)

const envVarRunVaultAccTests = "VAULT_ACC"

func TestVaultAccPathRole(t *testing.T) {
	t.Parallel()
	if os.Getenv(envVarRunVaultAccTests) != "1" {
		t.Skip("skipping end-to-end test (VAULT_ACC env var not set)")
	}

	envOptions := &stepwise.MountOptions{
		RegistryName:    "vault-artifactory-secrets-plugin",
		MountPathPrefix: "artifactory",
	}

	roleName := "ete_test_role"
	repo := envOrDefault("ARTIFACTORY_REPOSITORY_NAME", "ANY")
	rawPt := fmt.Sprintf(`
	[
		{
			"repo": {
				"include_patterns": ["/mytest/**"],
				"exclude_patterns": [""],
				"repositories": ["%s"],
				"operations": ["read", "write", "annotate"]
			}
		}
	]
	`, repo)

	stepwise.Run(t, stepwise.Case{
		Precheck:    func() { testVaultAccPreCheck(t) },
		Environment: newEteEnv(envOptions),
		// SkipTeardown: true,
		Steps: []stepwise.Step{
			testVaultAccConfig(t),
			testVaultAccRoleCreate(t, roleName, rawPt),
			testVaultAccRoleRead(t, roleName, rawPt),
			testVaultAccTokenUpdate(t, roleName),
		},
	})
}

var initSetup sync.Once

func testVaultAccPreCheck(t *testing.T) {
	initSetup.Do(func() {
		// Ensure Vault and Artifactory env variables are set
		if v := os.Getenv("ARTIFACTORY_URL"); v == "" {
			t.Fatal("ARTIFACTORY_URL not set")
		}
		if v := os.Getenv("ARTIFACTORY_BEARER_TOKEN"); v == "" {
			t.Fatal("ARTIFACTORY_BEARER_TOKEN not set")
		}
		if v := os.Getenv("VAULT_ADDR"); v == "" {
			t.Fatal("VAULT_ADDR not set")
		}
		if v := os.Getenv("VAULT_TOKEN"); v == "" {
			t.Fatal("VAULT_TOKEN  not set")
		}
	})
}

func testVaultAccConfig(t *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      configPrefix,
		Data: map[string]interface{}{
			"base_url":     os.Getenv("ARTIFACTORY_URL"),
			"bearer_token": os.Getenv("ARTIFACTORY_BEARER_TOKEN"),
			"max_ttl":      "3600s",
		},
	}
}

func testVaultAccRoleCreate(t *testing.T, roleName, pt string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.WriteOperation,
		Path:      "roles/" + roleName,
		Data: map[string]interface{}{
			"permission_targets": pt,
			"name":               roleName,
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			return nil
		},
	}
}

func testVaultAccRoleRead(t *testing.T, roleName, pt string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "roles/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.Equal(t, roleName, resp.Data["name"])
			return nil
		},
	}
}

func testVaultAccTokenUpdate(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "token/" + roleName,
		Data: map[string]interface{}{
			"role_name": roleName,
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["access_token"])
			return nil
		},
	}
}

type vaultAccEnv struct {
	options   *stepwise.MountOptions
	client    *api.Client
	mountPath string
}

var _ stepwise.Environment = (*vaultAccEnv)(nil)

func newEteEnv(options *stepwise.MountOptions) stepwise.Environment {
	return &vaultAccEnv{
		options: options,
	}
}

// Setup creates the Vault client to use against the test instance, and mounts the plugin to a
// unique path.
func (e *vaultAccEnv) Setup() error {

	// uses DefaultConfig
	c, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	e.client = c

	err = e.client.Sys().Mount(e.MountPath(), &api.MountInput{
		Type: e.options.RegistryName,
	})
	return err
}

func (e *vaultAccEnv) Client() (*api.Client, error) {
	return e.client.Clone()
}

func (e *vaultAccEnv) Teardown() error {
	return e.client.Sys().Unmount(e.mountPath)
}

func (e *vaultAccEnv) MountPath() string {
	if e.mountPath != "" {
		return e.mountPath
	}

	uuidStr, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	e.mountPath = fmt.Sprintf("%s_%s", e.options.MountPathPrefix, uuidStr)
	return e.mountPath
}

func (e *vaultAccEnv) Name() string {
	return "docker"
}

func (e *vaultAccEnv) RootToken() string {
	return os.Getenv("VAULT_TOKEN")
}
