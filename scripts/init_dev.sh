#!/bin/bash

set -euo pipefail

export PATH=$(pwd)/.tools:$PATH

DIR=$(dirname "$0")

. $DIR/wait_for.sh

pushd $DIR &>/dev/null
# docker-compose version >&2

set +u
if [ -n "$CI" ]; then
  echo "Starting containers in net=host mode..." >&2
  docker-compose -f docker-compose-ci.yaml up -d >&2
else
  docker-compose up -d >&2
fi
set -u

wait_for vault >&2
wait_for artifactory >&2

popd &>/dev/null
echo "Configuring artifactory..." >&2
# eval to capture ARTIFACTORY_ env vars
eval $("$DIR/setup_dev_artifactory.sh")

# if local env, configure vault plugin to talk to artifactory container hostname
set +u
if [ -z "$CI" ] || [ -n "$VAULT_ACC" ]; then
  export ARTIFACTORY_URL='http://artifactory:8081/artifactory/'
fi

echo "Configuring vault..." >&2
# eval to capture VAULT_ env vars
eval $("$DIR/setup_dev_vault.sh")

# For end-to-end tests using Vault in a Docker container, VAULT_ACC will be set and ARTIFACTORY_URL
# will refer to container hostname.
# For acc tests using vault test backend (non-docker), ARTIFACTORY_URL will be localhost.
if [ -z "$VAULT_ACC" ]; then
  export ARTIFACTORY_URL='http://localhost:8081/artifactory/'
fi
# eval output for local use
echo -e "\n\033[1;33meval this script to set Artifactory/Vault env vars\033[0m\n" >&2
echo export ARTIFACTORY_USER=\"$ARTIFACTORY_USER\"\;
echo export ARTIFACTORY_PASSWORD=\"$ARTIFACTORY_PASSWORD\"\;
echo export ARTIFACTORY_URL=\"$ARTIFACTORY_URL\"\;
echo export ARTIFACTORY_BEARER_TOKEN=\"$ARTIFACTORY_BEARER_TOKEN\"\;
echo export ARTIFACTORY_API_KEY=\"$ARTIFACTORY_API_KEY\"\;
echo export VAULT_ADDR=\"$VAULT_ADDR\"\;
echo export VAULT_TOKEN=\"$VAULT_TOKEN\"\;

echo -e "\nExample usage to test plugin:" >&2
echo -e "\033[0;32mvault write artifactory-cloud/roles/role1 token_ttl=600 permission_targets=@scripts/sample_permission_targets.json\033[0m\n" >&2
