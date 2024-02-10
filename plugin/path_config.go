// Copyright  2024 Splunk, Inc.
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

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// schema for the configuring artifactory secrets plugin, this will map the fields coming in from the
// vault request field map
var configSchema = map[string]*framework.FieldSchema{
	"base_url": {
		Type:        framework.TypeString,
		Description: `Artifactory base url. e.g. https://myjfrog.example.com/artifactory/`,
	},
	"bearer_token": {
		Type:        framework.TypeString,
		Description: `Artifactory token that has permissions to generate other tokens`,
	},
	"username": {
		Type:        framework.TypeString,
		Description: `Artifactory user that has permissions to generate other tokens`,
	},
	"password": {
		Type:        framework.TypeString,
		Description: `Artifactory password associated with username`,
	},
	"max_ttl": {
		Type:        framework.TypeDurationSecond,
		Description: "Maximum time a token generated will be valid for. If <= 0, will use system default(3600).",
		Default:     3600,
	},
	"client_timeout": {
		Type:        framework.TypeDurationSecond,
		Description: "Artifactory HTTP client timeout at Transport layer. If <=0, will use system default(30).",
		Default:     30,
	},
}

func (backend *ArtifactoryBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"base_url":       cfg.BaseURL,
			"max_ttl":        int64(cfg.MaxTTL / time.Second),
			"client_timeout": int64(cfg.ClientTimeout / time.Second),
		},
	}, nil
}

func (backend *ArtifactoryBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := backend.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &ConfigStorageEntry{}
	}

	if baseURL, ok := data.GetOk("base_url"); ok {
		cfg.BaseURL = baseURL.(string)
	}

	if bearerToken, ok := data.GetOk("bearer_token"); ok {
		cfg.BearerToken = bearerToken.(string)
	}

	if username, ok := data.GetOk("username"); ok {
		cfg.Username = username.(string)
	}

	if password, ok := data.GetOk("password"); ok {
		cfg.Password = password.(string)
	}

	maxTTLRaw, ok := data.GetOk("max_ttl")
	if ok && maxTTLRaw.(int) > 0 {
		cfg.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	} else if cfg.MaxTTL == time.Duration(0) {
		cfg.MaxTTL = time.Duration(configSchema["max_ttl"].Default.(int)) * time.Second
	}

	clientTimeoutRaw, ok := data.GetOk("client_timeout")
	if ok && clientTimeoutRaw.(int) > 0 {
		cfg.ClientTimeout = time.Duration(clientTimeoutRaw.(int)) * time.Second
	} else if cfg.ClientTimeout == time.Duration(0) {
		cfg.ClientTimeout = time.Duration(configSchema["client_timeout"].Default.(int)) * time.Second
	}

	entry, err := logical.StorageEntryJSON(configPrefix, cfg)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func pathConfig(b *ArtifactoryBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: configPrefix,
			Fields:  configSchema,

			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathConfigRead,
				logical.UpdateOperation: b.pathConfigWrite,
			},

			HelpSynopsis:    pathConfigHelpSyn,
			HelpDescription: pathConfigHelpDesc,
		},
	}

	return paths
}

const pathConfigHelpSyn = `
Configure the Artifactory backend.
`

const pathConfigHelpDesc = `
The Artifactory backend requires credentials for managing groups and permission targets 
and creating an access token for a group. This endpoint is used to configure
those credentials as well as default values for the backend in general.

If multiple credentials are provided, it takes precendence on following order. 
Bearer Token -> API Key -> Username/Password
`
