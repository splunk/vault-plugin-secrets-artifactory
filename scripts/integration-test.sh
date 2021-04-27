#!/bin/bash

# This script serves as entrypoint to the reapplied ci job docker container.
# see test-wrapper.sh

set -uo pipefail

echo "Current directory:"
pwd

set +u
if [ "$CI_DEBUG_TRACE" = true ]; then
  docker inspect $(docker ps -q)
fi
set -u

make integration-test report
exit_code=$?

# any additional logic/debugging can go here

go-go vault --logout

exit ${exit_code}
