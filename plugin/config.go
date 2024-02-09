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
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configPrefix = "config"
)

// ConfigStorageEntry structure represents the config as it is stored within vault
type ConfigStorageEntry struct {
	BaseURL       string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	BearerToken   string        `json:"bearer_token" structs:"bearer_token" mapstructure:"bearer_token"`
	Username      string        `json:"username" structs:"username" mapstructure:"username"`
	Password      string        `json:"password" structs:"password" mapstructure:"password"`
	MaxTTL        time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`
	ClientTimeout time.Duration `json:"client_timeout" structs:"client_timeout" mapstructure:"client_timeout"`
}

func (backend *ArtifactoryBackend) getConfig(ctx context.Context, s logical.Storage) (*ConfigStorageEntry, error) {
	var cfg ConfigStorageEntry
	cfgRaw, err := s.Get(ctx, configPrefix)
	if err != nil {
		return nil, err
	}
	if cfgRaw == nil {
		return nil, nil
	}

	if err := cfgRaw.DecodeJSON(&cfg); err != nil {
		return nil, err
	}

	return &cfg, err
}
