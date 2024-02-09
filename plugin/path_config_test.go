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
			"base_url":       "https://example.jfrog.io/",
			"bearer_token":   "mybearertoken",
			"client_timeout": "15s",
			"max_ttl":        "600s",
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url":       "https://example.jfrog.io/",
			"client_timeout": int64(15),
			"max_ttl":        int64(600),
		}

		testConfigRead(t, backend, reqStorage, expected)
		testConfigUpdate(t, backend, reqStorage, map[string]interface{}{
			"max_ttl": "50s",
		})

		expected["max_ttl"] = int64(50)
		testConfigRead(t, backend, reqStorage, expected)

		testConfigUpdate(t, backend, reqStorage, map[string]interface{}{
			"client_timeout": "20s",
		})

		expected["client_timeout"] = int64(20)
		testConfigRead(t, backend, reqStorage, expected)
	})

	t.Run("user_pwd", func(t *testing.T) {
		t.Parallel()

		backend, reqStorage := getTestBackend(t, true)

		testConfigRead(t, backend, reqStorage, nil)

		conf := map[string]interface{}{
			"base_url":       "https://example.jfrog.io/",
			"username":       "uname",
			"password":       "pwd",
			"client_timeout": "2m",
			"max_ttl":        "1h",
		}

		testConfigUpdate(t, backend, reqStorage, conf)

		expected := map[string]interface{}{
			"base_url":       "https://example.jfrog.io/",
			"client_timeout": int64(120),
			"max_ttl":        int64(3600),
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
