<!-- omit in toc -->
# vault-plugin-secrets-artifactory

>Successor of [vault-artifactory-token-plugin]. Planned to be open-sourced.

This is a backend plugin to be used with Vault. This plugin generates one-time access tokens.

[Design doc][design-doc]

- [Requirements](#requirements)
- [Getting Started](#getting-started)
  - [Usage](#usage)
- [Documents](#documents)
  - [Update Permission Targets](#update-permission-targets)
  - [Garbage Collection](#garbage-collection)
- [Development](#development)
  - [Tests](#tests)
- [Roadmap](#roadmap)

## Requirements

- Go: 1.6 or above
- **Artifactory: 6.6.0** or above for API V2 support.
- token with admin privileges to manage groups and permission targets and to create tokens

## Getting Started

This is a [Vault plugin] meant to work with Vault. This guide assumes you have already installed
Vault and have a basic understanding of how Vault works.

Otherwise, first read [how to get started with Vault][vault-getting-started].

To learn specifically about how plugins work, see documentation on [Vault
plugins][vault plugin].

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

when a role is created, it generates an artifactory group and supplied permission targets. To
achieve unique group and permission target names per role, it applies a UUID to each
role as `role_id` and appends it to the group and permission target names:

| Artifactory Object | format                                                                | example                                             |
| ------------------ | --------------------------------------------------------------------- | --------------------------------------------------- |
| Group              | `vault-plugin.<role_id>`                                              | `vault-plugin.9ace47f6-a205-11eb-8b68-acde48001122` |
| Permission Target  | `vault-plugin.<"name" field of supplied permission target>.<role_id>` | `npm-test.9ace47f6-a205-11eb-8b68-acde48001122`     |

Token is generated with a transient user and returned as key value pair:

| key          | value                            |
| ------------ | -------------------------------- |
| access_token | REDACTED_BEARER_TOKEN            |
| username     | `auto-vault-plugin-user.ci-role` |

username follows the format of `auto-vault-plugin-user.<role_name>`

### Update Permission Targets

List of permission targets can be supplied as a JSON string. Format of a permission target can be
found [here][permission-target-format].

To apply a dynamically created group to permission targets, you must use `VAULT_PLUGIN_OWN_ROLE`
as group name permission target group key. For example:

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

You have notified that `actions` from V2 permission target are swapped with `operations`. This is
because the `actions` field can contain users and other groups which are obsolete in this plugin.

To update permission targets for an existing role, please also supply existing permission
targets in order to preserve them in a role. Updating without supplying existing
permission targets registered to a role **will delete those existing permission targets**.

```sh
# To grab existing permission targets
$ vault read artifactory/roles/ci-role -format=json | jq -r .data.permission_targets > permission_targets.json
```

### Garbage Collection

To keep the isolation, artifactory groups and permission targets are not shared amongst different
roles. To this nature, it collects garbage when update/delete operation is performed on a role:

- removal of an artifactory group and permission targets when the corresponding role is removed
- removal of an artifactory permission target  when it's removed from the corresponding role

## Development

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

```

You can then create a role and issue a token following above usage.

### Tests

```sh
# run unit tests
make test

# run subset of tests
make test TESTARGS='-run=TestConfig'

# generate a code coverage report
make report
open coverage.html

# run integration tests (spins up local vault and artifactory instance)
make integration-test
```

## Roadmap

This plugin is being initially developed for an internal application at Splunk for suite of
ephemeral credentials. No user or service account should be able to publish production artifacts
with a static credential.

Merge requests, issues and comments are always welcome.

[design-doc]:https://docs.google.com/document/d/1lfWFeutKLKrS39qFHDMmTZba5-6j628irv8HNLpASfc/edit#
[permission-target-format]:https://www.jfrog.com/confluence/display/JFROG/Security+Configuration+JSON#SecurityConfigurationJSON-application/vnd.org.jfrog.artifactory.security.PermissionTargetV2+json
[vault-artifactory-token-plugin]: 
[vault-getting-started]:https://www.vaultproject.io/intro/getting-started/install.html
[vault plugin]:https://www.vaultproject.io/docs/internals/plugins.html
