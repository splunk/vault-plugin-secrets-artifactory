package artifactorysecrets

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestConfigBearerToken(t *testing.T) {
	b, reqStorage := getTestBackend(t)

	testConfigRead(t, b, reqStorage, nil)

	conf := map[string]interface{}{
		"base_url":     "https://example.jfrog.io/example",
		"bearer_token": "mybearertoken",
		"max_ttl":      "600s",
	}

	testConfigUpdate(t, b, reqStorage, conf)

	expected := map[string]interface{}{
		"base_url": "https://example.jfrog.io/example",
		"max_ttl":  int64(600),
	}

	testConfigRead(t, b, reqStorage, expected)
	testConfigUpdate(t, b, reqStorage, map[string]interface{}{
		"max_ttl": "50s",
	})

	expected["max_ttl"] = int64(50)
	testConfigRead(t, b, reqStorage, expected)
}

func TestConfigUserPwd(t *testing.T) {
	b, reqStorage := getTestBackend(t)

	testConfigRead(t, b, reqStorage, nil)

	conf := map[string]interface{}{
		"base_url": "https://example.jfrog.io/example",
		"username": "uname",
		"password": "pwd",
		"max_ttl":  "600s",
	}

	testConfigUpdate(t, b, reqStorage, conf)

	expected := map[string]interface{}{
		"base_url": "https://example.jfrog.io/example",
		"max_ttl":  int64(600),
	}

	testConfigRead(t, b, reqStorage, expected)
	testConfigUpdate(t, b, reqStorage, map[string]interface{}{
		"max_ttl": "50s",
	})

	expected["max_ttl"] = int64(50)
	testConfigRead(t, b, reqStorage, expected)
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configPrefix,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configPrefix,
		Storage:   s,
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp == nil && expected == nil {
		return
	}

	if resp.IsError() {
		t.Fatal(resp.Error())
	}

	if len(expected) != len(resp.Data) {
		t.Errorf("read data mismatch (expected %d values, got %d)", len(expected), len(resp.Data))
	}

	if !reflect.DeepEqual(expected, resp.Data) {
		t.Errorf(`expected data %v not equal actual %v"`, expected, resp.Data)
	}

	if t.Failed() {
		t.FailNow()
	}
}
