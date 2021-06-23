#!/bin/bash

# This script serves as entrypoint to the reapplied ci job docker container.
# see test-wrapper.sh

set -uo pipefail

echo "Current directory:"
pwd

make ete-test
exit_code=$?

# any additional logic/debugging can go here

go-go vault --logout

exit ${exit_code}
