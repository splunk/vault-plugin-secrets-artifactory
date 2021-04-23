package artifactorysecrets

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// basic schema for the creation of the token,
// this will map the fields coming in from the vault request field map
var createTokenSchema = map[string]*framework.FieldSchema{
	"role_name": {
		Type:        framework.TypeString,
		Description: "The name of the role for which token is to be created",
	},
	"ttl": {
		Type:        framework.TypeDurationSecond,
		Description: "The duration in seconds after which the token will expire",
		Default:     600, // default of 10 minutes
	},
}

// create the basic jwt token with an expiry within the claim
func (backend *ArtifactoryBackend) createToken(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	roleName := data.Get("role_name").(string)

	roleEntry, err := backend.getRoleEntry(ctx, req.Storage, roleName)
	if roleEntry == nil || err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Role name '%s' not recognised", roleName)), err
	}

	var tokenEntry TokenCreateEntry

	if ttlRaw, ok := data.GetOk("ttl"); ok {
		tokenEntry.TTL = time.Duration(ttlRaw.(int)) * time.Second
	} else {
		tokenEntry.TTL = roleEntry.TokenTTL
	}

	if tokenEntry.TTL > roleEntry.MaxTTL {
		return logical.ErrorResponse(fmt.Sprintf("Token ttl is greater than role max ttl '%d'", roleEntry.MaxTTL)), err
	}

	token, err := backend.createTokenEntry(ctx, req.Storage, tokenEntry, roleEntry)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Error creating token, %#v", err)), err
	}

	return &logical.Response{Data: token}, nil
}

func pathToken(backend *ArtifactoryBackend) []*framework.Path {
	paths := []*framework.Path{
		{
			Pattern: fmt.Sprintf("token/%s", framework.GenericNameRegex("role_name")),
			Fields:  createTokenSchema,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: backend.createToken,
				logical.UpdateOperation: backend.createToken,
			},
			HelpSynopsis:    pathTokenHelpSyn,
			HelpDescription: pathTokenHelpDesc,
		},
	}

	return paths
}

const pathTokenHelpSyn = `Generate an access token under a specific role.`
const pathTokenHelpDesc = `
This path will generate a new access token for accessing Artifactory APIs.
A role, binding permission targets to specific Artifactory resources, will be specified
by name - for example, if this backend is mounted at "artifactory",
then "artifactory/token/deploy" would generate tokens for the "deploy" role.

On the backend, each role is associated with a group.
The token will be associated with this group. Tokens have a
short-term lease (default 10-mins) associated with them but cannot be renewed.
`
