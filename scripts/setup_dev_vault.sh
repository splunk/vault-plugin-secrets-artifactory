#!/bin/bash

vault secrets enable -path=artifactory vault-artifactory-secrets-plugin
vault write artifactory/config base_url=$ARTIFACTORY_URL bearer_token=$BEARER_TOKEN ttl=600 max_ttl=600
#vault write artifactory/roles/role1 token_ttl=600 permission_targets=@sample_permission_targets.json
