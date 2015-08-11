#!/bin/bash

"${WERCKER_STEP_ROOT}/wercker-step-distelli"
STATUS=$?
if [ ${STATUS} -ne 0 ]; then
  echo "Failed executing distelli step." >&2
  exit 1
fi
