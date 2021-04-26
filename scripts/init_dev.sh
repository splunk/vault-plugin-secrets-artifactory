#!/bin/bash

set -euo pipefail

export PATH=$(pwd)/.tools:$PATH

DIR=$(dirname "$0")

. $DIR/wait_for.sh

pushd $DIR &>/dev/null
docker-compose version >&2

set +u
if [ -n "$CI" ]; then
  echo "Starting containers in net=host mode..."
  docker-compose -f docker-compose-ci.yaml up -d
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
if [ -z "$CI" ]; then
  export ARTIFACTORY_URL='http://artifactory:8081/artifactory'
fi

echo "Configuring vault..." >&2
# eval to capture VAULT_ env vars
eval $("$DIR/setup_dev_vault.sh")

if [ -z "$CI" ]; then
  # ARTIFACTORY_URL will always be localhost outside of containers
  export ARTIFACTORY_URL='http://localhost:8081/artifactory'
  # eval output for local use
  echo -e "\n\033[1;33mCopy/paste or eval this script:\033[0m\n" >&2
  echo export ARTIFACTORY_USER=\"$ARTIFACTORY_USER\"\;
  echo export ARTIFACTORY_PASSWORD=\"$ARTIFACTORY_PASSWORD\"\;
  echo export ARTIFACTORY_URL=\"$ARTIFACTORY_URL\"\;
  echo export ARTIFACTORY_BEARER_TOKEN=\"$ARTIFACTORY_BEARER_TOKEN\"\;
  echo export VAULT_ADDR=\"$VAULT_ADDR\"\;
  echo export VAULT_TOKEN=\"$VAULT_TOKEN\"\;
fi
set -u

echo -e "\nExample usage to test plugin:" >&2
echo -e "\033[0;32mvault write artifactory/roles/role1 token_ttl=600 permission_targets=@scripts/sample_permission_targets.json\033[0m\n" >&2
