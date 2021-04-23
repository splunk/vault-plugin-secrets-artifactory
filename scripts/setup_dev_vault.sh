#!/bin/bash

: "${ARTIFACTORY_URL:?unset}"

vault secrets enable -path=artifactory vault-artifactory-secrets-plugin

if [ -z "$BEARER_TOKEN" ]; then
  echo "BEARER_TOKEN unset, using USERNAME/PASSWORD"
  : "${USERNAME:?unset}"
  : "${PASSWORD:?unset}"
  vault write artifactory/config base_url=$ARTIFACTORY_URL username=$USERNAME password=$PASSWORD ttl=600 max_ttl=600
else
  vault write artifactory/config base_url=$ARTIFACTORY_URL bearer_token=$BEARER_TOKEN ttl=600 max_ttl=600
fi

vault write artifactory/roles/role1 token_ttl=600 permission_targets=@$(dirname $0)/sample_permission_targets.json
