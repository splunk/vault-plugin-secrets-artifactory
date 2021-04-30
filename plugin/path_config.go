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
		Description: `Artifactory base url. e.g. htts://myjfrog.example.com/artifactory/`,
	},
	"bearer_token": {
		Type:        framework.TypeString,
		Description: `Artifactory token that has permissions to generate other tokens`,
	},
	"api_key": {
		Type:        framework.TypeString,
		Description: `Artifactory API key of a user that has permissions to generate other tokens`,
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
			"base_url": cfg.BaseURL,
			"max_ttl":  int64(cfg.MaxTTL / time.Second),
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
		url := appendTrailingSlash(baseURL.(string))
		cfg.BaseURL = url
	}

	if bearerToken, ok := data.GetOk("bearer_token"); ok {
		cfg.BearerToken = bearerToken.(string)
	}

	if apiKey, ok := data.GetOk("api_key"); ok {
		cfg.ApiKey = apiKey.(string)
	}

	if username, ok := data.GetOk("username"); ok {
		cfg.Username = username.(string)
	}

	if password, ok := data.GetOk("password"); ok {
		cfg.Password = password.(string)
	}

	if maxTTLRaw, ok := data.GetOk("max_ttl"); ok {
		cfg.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	} else {
		cfg.MaxTTL = time.Duration(configSchema["max_ttl"].Default.(int)) * time.Second
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
