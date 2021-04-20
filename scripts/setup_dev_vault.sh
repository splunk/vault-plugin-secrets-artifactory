#!/bin/bash

vault secrets enable -path=artifactory-cloud vault-artifactory-secrets-plugin
vault write artifactory-cloud/config base_url=$ARTIFACTORY_URL bearer_token=$BEARER_TOKEN ttl=600 max_ttl=600
# cat ./scripts/sample_permission_targets.json | vault write artifactory-cloud/roles/role1 token_ttl=600 permission_targets=-
