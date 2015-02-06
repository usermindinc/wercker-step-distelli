#!/bin/bash

# Run in a subshell to prevent corrupting the build environment
(
source "${WERCKER_STEP_ROOT}/setup.sh"

python "${WERCKER_STEP_ROOT}/run.py"
)

