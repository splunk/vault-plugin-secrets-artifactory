# vault-plugin-secrets-artifactory

This is a backend plugin to be used with Vault. This plugin generates one-time access tokens.

## Requirements

Go: 1.6 or above
Artifactory: 6.6.0 or above for API V2 support
A token with admin privileges to manage groups and permission targets and to create tokens

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/internals/plugins.html).

### Usage

```sh
# Please mount a plugin, then you can enable a secret
$ vault secrets enable -path=artifactory vault-artifactory-secrets-plugin
Success! Enabled the vault-artifactory-secrets-plugin secrets engine at: artifactory/

# configure the /config backend. You must supply admin bearer token or username/password pair of an admin user.
$ vault write artifactory/config base_url="https://artifactory.example.com/artifactory" bearer_token=$BEARER_TOKEN ttl=600 max_ttl=600

# creating a role
$ vault write artifactory/roles/ci-role token_ttl=600 permission_targets=@scripts/sample_permission_targets.json

$ vault write artifactory/token/ci-role ttl=60
Key             Value
---             -----
access_token    REDACTED
username        auto-vault-plugin-user.ci-role
```

## Documents

when a role is created, it generates an artifactory group and supplied permission targets.
To achieve uniqueness per role in artifactory group and permission target, it applies UUID to each role as `role_id` and append `role_id` to group and permission targets name in following rule
| Artifactory Object | format | example |
| --- | --- | --- |
| Group | `vault-plugin.<role_id>` | `vault-plugin.9ace47f6-a205-11eb-8b68-acde48001122` |
| Permission Target | `vault-plugin.<"name" field of supplied permission target>.<role_id>` | `npm-test.9ace47f6-a205-11eb-8b68-acde48001122` |

Token is generated with a transient user and returned as key value pair. 
| key | value |
| -- | -- |
| access_token | REDUCTED_BEAERER_TOKEN |
| username | `auto-vault-plugin-user.ci-role` |

username follows the format of `auto-vault-plugin-user.<role_name>`

### Update Permission Targets

List of permission targets can be supplied as JSON string.

```json
[
  {
    "name": "docker",
    "repo": {
      "include_patterns": ["/myprefix/**", "/anotherprefix/myteam/**"] ,
      "exclude_patterns": [""],
      "repositories": ["docker-local"],
      "operations": ["read"]
    }
  },
]
```

You have notified that `actions` from V2 permission target are swapped with `operations`. This is because the `actions` field can contain users and other groups which are obsolete in this plugin.  

To update permission targets for an existing role, please also supply existing permisssion targets in order to preserve them in a role. Update operation without supplying existing permission targets registered to a role will delete those existing permission targets

```sh
# To grab existing permission targets
$ vault read artifactory/roles/ci-role -format=json | jq -r .data.permission_targets > permission_targets.json
```

### Garbage Collection

To keep the isolation, it doesn't share an artifactory group or permission targets amongst different roles. To this nature, it collects garbage where it's unnecessary when update/delete operation is performed on a role

- removal of an artifactory group and permission targets when the corresponding role is removed
- removal of an artifactory permission target  when it's removed from the corresponding role

## Testing Locally

Requirements:

- vault

```sh
# Build binary in plugins directory
$ make build

# Start vault dev server
$ make dev-server

# New terminal
$ export VAULT_ADDR=https://127.0.0.1:8200
$ export VAULT_TOKEN=root
$ export ARTIFACTORY_URL="https://artifactory.example.com/artifactory"
$ export BEARER_TOKEN=TOKEN

# enable secrets backend and configuration
$ ./scripts/setup_dev_vault.sh

# You can then create a role and issue a token following above usage. 
```

## Tests

To run unit tests,

```sh
$ make test
```

To run integration tests that spins up local vault and artifactory instance, 

```sh
$ make integration-test
```

## Roadmap

This plugin is being initially developed for an internal application at Splunk for suite of ephemeral credentials. No user or service account should be able to publish production artifacts with a static credential.
