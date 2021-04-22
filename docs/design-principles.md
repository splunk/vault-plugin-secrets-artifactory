# Artifactory Secrets Engine

The Artifactory Vault secrets engine dynamically generates artifactory access token based on permission targets. This enables users to gain access to Artifactory repositories without needing to create or manage a dedicated service account.  

## Design Principles

This plugin is influenced by [Google Cloud Secrets Engine](https://www.vaultproject.io/docs/secrets/gcp)

a bunch of paths and repositories on which people can read or write or both, say "permission targets set". The set is treated as one or more of permission targets. We bind these permission targets to a role, aka group in artifactory term. When a token is acquired, it is scoped to the group.  

- one artifactory group per role
- multiple permission targets per role
- no sharing of permission targets amongst roles
- garbage collection - artifactory group and permission targets

path /roles/<role_name>  

- Get triggers ‘get group name’ and it should come with permission targets
- Create/Update triggers group and permission targets creation
  - Update include deletion of permission targets if it’s removed from bindings
- Delete triggers deletion of group and permission targets

path /token/<role_name>

- Create/Update triggers token creation scoped to the group

## Things to Note

### Roles Are Tied to Permission Targets

Artifactory groups are created when a role is created rather than each time a secret is generated. This may different from how other secrets engines behave, but it is for a good reasons:

- Artifactory Group and Permission Targets creation can be limited by many factors such as Artifactory API, repositories existence and database limitation on length of group/permission targets. By creating the group and permission targets in advance, we can speed up the timeliness of future operations and reduce the flakiness of automated workflows

### Respositories Must Exist at Role Creation

Because the permission targets for the group are set during role creation, repositories that do not exist will fail the `Create or Replace Permission Target` API call.

### Role Creation May Partially Fail

Every group creation and permission target creation is an Artifactory API call per resource. If an API call to one of these resources fails, the role creation fails and Vault will attempt to rollback.

These rollbackls are API calls, so they may also fail. The secrets engine uses WAL to ensure that unused permission targets are cleaned up. In the case of api failures, you may need to clean these up manually.

### Do Not Modify Vault-owned Group and Permission Targets

While Vault will initially create and assign permission targets to groups, it is possible that an external user deletes or modifies this group and/or permission targets. These changesare difficult to detect, and it is best to prevent this type of modification.  

Vault-owned group have in the format: `vault-plugin.<UUID of Role ID>`
Vault-owned permission target have in the format: `vault-plugin.pt<index of permission target counts>.<UUID of Role ID>`

Communicate with your teams to not modify these resources.
