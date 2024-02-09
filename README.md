<!-- omit in toc -->
# vault-plugin-secrets-artifactory

[![build-status-badge]][actions-page]
[![go-report-card-badge]][go-report-card]
[![codecov-badge]][codecov]
![go-version-badge]

This is a backend plugin to be used with Vault. This plugin generates one-time access tokens.

[Design doc][design-doc]

- [Requirements](#requirements)
- [Getting Started](#getting-started)
  - [Usage](#usage)
- [Documents](#documents)
  - [Update Permission Targets](#update-permission-targets)
  - [Garbage Collection](#garbage-collection)
- [Development](#development)
  - [Full dev environment](#full-dev-environment)
  - [Developing with an existing Artifactory instance](#developing-with-an-existing-artifactory-instance)
  - [Tests](#tests)
- [License](#license)

## Requirements

- Go: 1.22 or above
- **Artifactory: 7.21.1** or above (for Access API support)
- **Artifactory Pro or above is required** for the [API endpoints][artifactory-api-ref] used by
  this plugin. A license key will be needed to spin up the full dev environment.
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
# URL can have /artifactory/ but this will be stripped for the Access API (`/access/`).
$ vault write artifactory/config base_url="https://artifactory.example.com/artifactory" bearer_token=$BEARER_TOKEN ttl=600 max_ttl=600

# see supported paths
$ vault path-help artifactory/
$ vault path-help artifactory/config

# create a role
$ vault write artifactory/roles/ci-role token_ttl=600 permission_targets=@scripts/sample_permission_targets.json

# generate an ephemeral artifactory token
$ vault write artifactory/token/ci-role ttl=60
Key             Value
---             -----
access_token    REDACTED
username        auto-vault-plugin-user.ci-role
```


**Note:** If username/password is used, [Enable Token Generation via
API](https://jfrog.com/help/r/jfrog-platform-administration-documentation/enable-token-generation-via-api)
is required to be set in the Artifactory instance.

## Documents

when a role is created, it generates an artifactory group and supplied permission targets. To
achieve unique group and permission target names per role, it applies a UUID to each
role as `role_id` and appends it to the group and permission target names:

| Artifactory Object | format                                                           | example                                             |
| ------------------ | ---------------------------------------------------------------- | --------------------------------------------------- |
| Group              | `vault-plugin.<role_id>`                                         | `vault-plugin.9ace47f6-a205-11eb-8b68-acde48001122` |
| Permission Target  | `vault-plugin.pt<index of permission target counts>.<role_name>` | `npm-test.pt0.ci-role`                              |

Group name uses UUID as it's bounded to max 64 chars DB limit, whereas permission target name can be
longer than that.

Token is generated with a transient user and returned as key value pair:

| key          | value                            |
| ------------ | -------------------------------- |
| access_token | REDACTED_BEARER_TOKEN            |
| username     | `auto-vault-plugin-user.ci-role` |

username follows the format of `auto-vault-plugin-user.<role_name>`  
*note: if role name exceeds 39 characters, it shortens to fit into max char constraints*

### Update Permission Targets

List of permission targets can be supplied as a JSON string. Format of a permission target can be
found below. This is derived from artifactory V2 security permission target json that you can find
[here][permission-target-format].

```json
[
  {
    "repo": {
      "include_patterns": ["/myprefix/**", "/myteam/anotherprefix/**"] ,
      "exclude_patterns": [""],
      "repositories": ["docker-local"],
      "operations": ["read"]
    }
  },
  {
    "build": {
      "include_patterns": [""] ,
      "exclude_patterns": [""],
      "repositories": ["artifactory-build-info"],
      "operations": ["read"]
    }
  },
]
```

You have noticed that `actions` from V2 permission target are swapped with `operations`. This is
because the `actions` field can contain users and other groups which are obsolete in this plugin.

To update permission targets for an existing role, please also supply existing permission
targets in order to preserve them in a role. Updating without supplying existing
permission targets registered to a role **will delete those existing permission targets**.

```sh
# To grab existing permission targets
$ vault read artifactory/roles/ci-role -format=json | jq '.data.permission_targets|fromjson' > permission_targets.json
```

### Garbage Collection

To keep the isolation, artifactory groups and permission targets are not shared amongst different
roles. To this nature, it collects garbage when update/delete operation is performed on a role:

- removal of an artifactory group and permission targets when the corresponding role is removed
- removal of an artifactory permission target  when it's removed from the corresponding role

## Development

### Full dev environment

This will spin up an Artifactory Pro instance and Vault server in dev mode with the plugin
configured.

Requirements:

- docker

```sh
export ARTIFACTORY_LICENSE_KEY="<licenseKey>"

# spin up dev environment and print out env vars necessary for Vault/Artifactory.
make dev

# or do this to capture capture Artifactory/Vault env vars:
make tools
eval $(make dev)
```

To access the dev env Artifactory UI, navigate to [http://localhost:8082](http://localhost:8082)
and log in with the `ARTIFACTORY_USER` and `ARTIFACTORY_PASSWORD` output above.

### Developing with an existing Artifactory instance

Requirements:

- vault

```sh
# Build binary in plugins directory
make build

# Start a standalone vault dev server
make vault-only

# New terminal
export VAULT_ADDR=http://localhost:8200
export ARTIFACTORY_URL="https://artifactory.example.com/artifactory/"
export ARTIFACTORY_BEARER_TOKEN=TOKEN

# enable secrets backend and configuration
./scripts/setup_dev_vault.sh

```

You can then create a role and issue a token following above usage.

### Tests

```sh
# run unit tests
make test

# run subset of tests
make test TESTARGS='-run=TestConfig'

# run Artifactory acceptance tests (uses in-memory vault backend with Artifactory Docker container)
make test-artacc

# run Vault acceptance tests (uses Vault and Artifactory Docker containers against the compiled plugin)
make test-vaultacc

# generate a code coverage report
make report
open coverage.html
```

## License

[Apache Software License version 2.0](LICENSE)

[actions-page]:https://github.com/splunk/vault-plugin-secrets-artifactory/actions
[artifactory-api-ref]:https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API
[build-status-badge]:https://github.com/splunk/vault-plugin-secrets-artifactory/workflows/test.yml/badge.svg
[codecov]:https://codecov.io/gh/splunk/vault-plugin-secrets-artifactory
[codecov-badge]:https://codecov.io/gh/splunk/vault-plugin-secrets-artifactory/branch/main/graph/badge.svg
[design-doc]:https://docs.google.com/document/d/1lfWFeutKLKrS39qFHDMmTZba5-6j628irv8HNLpASfc/edit#
[go-report-card]:https://goreportcard.com/report/github.com/splunk/vault-plugin-secrets-artifactory
[go-report-card-badge]:https://goreportcard.com/badge/github.com/splunk/vault-plugin-secrets-artifactory
[go-version-badge]:https://img.shields.io/github/go-mod/go-version/splunk/vault-plugin-secrets-artifactory
[permission-target-format]:https://www.jfrog.com/confluence/display/JFROG/Security+Configuration+JSON#SecurityConfigurationJSON-application/vnd.org.jfrog.artifactory.security.PermissionTargetV2+json
[vault-getting-started]:https://www.vaultproject.io/intro/getting-started/install.html
[vault plugin]:https://www.vaultproject.io/docs/internals/plugins.html
