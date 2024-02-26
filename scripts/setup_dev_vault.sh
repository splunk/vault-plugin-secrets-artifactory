#!/bin/bash

set -eo pipefail

: "${ARTIFACTORY_URL:?unset}"

export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN=root

setup_vault() {
  plugin=vault-artifactory-secrets-plugin
  existing=$(vault secrets list -format json | jq -r '."artifactory-cloud/"')
  if [ "$existing" == "null" ]; then

    # in CI, current container bind mount is rprivate, preventing nested bind mounts
    # instead, copy plugin in to vault container and reload
    if ! vault plugin list secret | grep -q artifactory; then
      echo "Plugin missing from dev plugin dir /vault/plugins... registering manually."
      sha=$(shasum -a 256 plugins/$plugin | cut -d' ' -f1)
      # if plugin is missing, it is assumed this is a CI environment and vault is running in a container
      docker cp plugins/$plugin vault:/vault/plugins
      vault plugin register -sha256="$sha" secret "$plugin"
    fi

    echo "Enabling vault artifactory plugin..."
    vault secrets enable -path=artifactory-cloud $plugin

  else
    echo
    echo  "Plugin enabled on path 'artifactory-cloud/':"
    echo "$existing" | jq
    sha=$(shasum -a 256 plugins/$plugin | cut -d' ' -f1)
    vault plugin register -sha256="$sha" secret "$plugin"
    vault plugin reload -plugin=$plugin
  fi

  if [ -z "$ARTIFACTORY_BEARER_TOKEN" ]; then
    echo "ARTIFACTORY_BEARER_TOKEN unset, using username/password"
  : ${ARTIFACTORY_USER:?unset}
  : ${ARTIFACTORY_PASSWORD:?unset}
    vault write artifactory-cloud/config base_url=$ARTIFACTORY_URL username=$ARTIFACTORY_USER password=$ARTIFACTORY_PASSWORD ttl=600 max_ttl=600
  else
    vault write artifactory-cloud/config base_url=$ARTIFACTORY_URL bearer_token=$ARTIFACTORY_BEARER_TOKEN ttl=600 max_ttl=24h
  fi
}

setup_vault >&2

# eval output for local use
echo export VAULT_ADDR=\"$VAULT_ADDR\"\;
echo export VAULT_TOKEN=\"$VAULT_TOKEN\"\;
