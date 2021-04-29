package artifactorysecrets

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("bearer_token", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url":     "https://example.jfrog.io/example",
			"bearer_token": "mybearertoken",
			"max_ttl":      "600s",
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url": "https://example.jfrog.io/example/",
			"max_ttl":  int64(600),
		}

		testConfigRead(t, backend, reqStorage, expected)
		testConfigUpdate(t, backend, reqStorage, map[string]interface{}{
			"max_ttl": "50s",
		})

		expected["max_ttl"] = int64(50)
		testConfigRead(t, backend, reqStorage, expected)
	})

	t.Run("api_key", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url": "https://example.jfrog.io/example/",
			"api_eky":  "myapikey",
			"max_ttl":  "300s",
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url": "https://example.jfrog.io/example/",
			"max_ttl":  int64(300),
		}

		testConfigRead(t, backend, reqStorage, expected)
	})

	t.Run("user_pwd", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url": "https://example.jfrog.io/example",
			"username": "uname",
			"password": "pwd",
			"max_ttl":  "1h",
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url": "https://example.jfrog.io/example/",
			"max_ttl":  int64(3600),
		}

		testConfigRead(t, backend, reqStorage, expected)
	})
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configPrefix,
		Data:      d,
		Storage:   s,
	})
	require.NoError(t, err)
	require.False(t, resp.IsError())
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configPrefix,
		Storage:   s,
	})

	require.NoError(t, err)

	if resp == nil && expected == nil {
		return
	}

	require.False(t, resp.IsError())
	assert.Equal(t, len(expected), len(resp.Data), "read data mismatch")
	assert.Equal(t, expected, resp.Data, "expected %v, actual: %v", expected, resp.Data)

	if t.Failed() {
		t.FailNow()
	}
}
