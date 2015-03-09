#!/bin/bash

distelli="${WERCKER_STEP_ROOT}/DistelliCLI/bin/distelli"

check_branches() {
  local branches=${WERCKER_DISTELLI_BRANCHES}

  if [ ! -n "${branches}" ]; then
    return 0
  fi

  local IFS=","

  for b in ${branches}
  do
    if [ "$b" = "${WERCKER_GIT_BRANCH}" ]; then
      return 0
    fi
  done

  info "Current branch ${WERCKER_GIT_BRANCH} not in permitted set ${branches}, skipping."
  exit 0
}

check_manifest() {

  local manifest=${WERCKER_DISTELLI_MANIFEST}

  if [ ! -n "${manifest}" ]; then
    fail "manifest must be set"
  fi

  if [ ! -f "${manifest}" ]; then
    fail "manifest file ${manifest} not found"
  fi

  echo "${manifest}"

}

locate_app_name() {
  local app=${WERCKER_DISTELLI_APPLICATION}
  if [ ! -n "${app}" ]; then
    # TODO parse manifest file for application name
    fail "For now, app name has to be specified on deploy"
    return $?
  fi

  echo "${app}"
}

locate_release_id() {

  local release_id=${WERCKER_DISTELLI_RELEASE}

  if [ ! -n "${release_id}" ]; then
    # Nothing was specified, so we need to query distelli and look for the release
    local app=$(locate_app_name)
    status=$?

    if [ $status -ne 0 ]; then
      fail "Unable to locate application name for deployment"
    fi

    ${distelli} list releases -n ${app} -f csv | while read line; do
      IFS=',' read -a release_data <<< $line
      # Array columns are name, release id, created, description
      description=${release_data[3]}
      if [ "${description}" = "wercker:${WERCKER_BUILD_ID}" ]; then
        release_id=${release_data[1]}
        break
      fi
    done
  fi

  echo "$release_id"

}

push() {

  local manifest=$(check_manifest)

  # Distelli 1.88 assumes manifest is in CWD

  local dirname=$(dirname "${manifest}")
  local basename=$(basename "${manifest}")
  pushd "${dirname}"

  # Wercker checks us out to a commit, not a branch name (sensible, since the
  # branch may have moved on). Distelli doesn't handle this well. We won't have
  # any local branches (except master), so create one with an appropriate name.
  
  git checkout -b "wercker-${WERCKER_GIT_BRANCH}-${WERCKER_GIT_COMMIT:0:7}"

  echo "${distelli}" push -f "${basename}" -m "wercker:${WERCKER_BUILD_ID}"
  echo "${distelli}" bundle -f "${basename}" -b "${WERCKER_OUTPUT_DIR}"
  "${distelli}" bundle -f "${basename}" -b "${WERCKER_STEP_TMP}"

  popd

}

deploy() {

  local args=()

  local manifest=$(check_manifest)

  local environment=${WERCKER_DISTELLI_ENVIRONMENT}

  local host=${WERCKER_DISTELLI_HOST}

  if [ -n "${environment}" ]; then
    if [ -n "${host}" ]; then
      fail "Both environment and host are set"
    fi  
    args+=("-e" "${environment}")
  elif [ -n "${host}" ]; then
    args+=("-h" "${host}")
  else
    fail "Either environment or host must be set"
  fi
  
  # Distelli 1.88 assumes manifest is in CWD

  local dirname=$(dirname "${manifest}")
  local basename=$(basename "${manifest}")
  pushd "${dirname}"

  args+=("-f" "${basename}")

  info "wut"

  release_id=$(locate_release_id)

  info "wat"

  if [ $? -ne 0 ]; then
    fail "Failure locating release id"
  fi

  if [ -n "${release_id}" ]; then
    args+=("-r" "${release_id}")
  fi

  echo "${distelli}" deploy "${args[@]}"

  popd
}

main() {
  "${distelli}" version

  check_branches

  command=${WERCKER_DISTELLI_COMMAND}

  if [ ! -n "${command}" ]; then
    fail "command must be set"
  fi

  case ${command} in
    push)
        push
	;;
    deploy)
        deploy
	;;
    *)
        fail "unknown command: ${command}"
	;;
  esac
}

main
