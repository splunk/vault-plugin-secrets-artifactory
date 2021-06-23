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

const envVarRunEteTests = "VAULT_ACC"

func TestEtePathRole(t *testing.T) {
	t.Parallel()
	if os.Getenv(envVarRunEteTests) != "1" {
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
		Precheck:    func() { testEtePreCheck(t) },
		Environment: newEteEnv(envOptions),
		// SkipTeardown: true,
		Steps: []stepwise.Step{
			testEteConfig(t),
			testEteRoleCreate(t, roleName, rawPt),
			testEteRoleRead(t, roleName, rawPt),
			testEteTokenUpdate(t, roleName),
		},
	})
}

var initSetup sync.Once

func testEtePreCheck(t *testing.T) {
	initSetup.Do(func() {
		// Ensure Vault and Artifactory env variables are set
		if v := os.Getenv("ARTIFACTORY_URL"); v == "" {
			t.Fatal("ARTIFACTORY_BEARER_TOKEN not set")
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

func testEteConfig(t *testing.T) stepwise.Step {
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

func testEteRoleCreate(t *testing.T, roleName, pt string) stepwise.Step {
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

func testEteRoleRead(t *testing.T, roleName, pt string) stepwise.Step {
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

func testEteTokenUpdate(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "token/" + roleName,
		Data: map[string]interface{}{
			"role_name": roleName,
		},
		Assert: func(resp *api.Secret, err error) error {
			// fmt.Printf("resp:\n %+v", resp)
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["access_token"])
			return nil
		},
	}
}

type eteEnv struct {
	options   *stepwise.MountOptions
	client    *api.Client
	mountPath string
}

var _ stepwise.Environment = (*eteEnv)(nil)

func newEteEnv(options *stepwise.MountOptions) stepwise.Environment {
	return &eteEnv{
		options: options,
	}
}

// Setup creates the Vault client to use against the test instance, and mounts the plugin to a
// unique path.
func (e *eteEnv) Setup() error {

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

func (e *eteEnv) Client() (*api.Client, error) {
	return e.client.Clone()
}

func (e *eteEnv) Teardown() error {
	return e.client.Sys().Unmount(e.mountPath)
}

func (e *eteEnv) MountPath() string {
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

func (e *eteEnv) Name() string {
	return "docker"
}

func (e *eteEnv) RootToken() string {
	return os.Getenv("VAULT_TOKEN")
}
