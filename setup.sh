#!/bin/bash

export PYENV_ROOT="${WERCKER_CACHE_DIR}/.pyenv"
PYENV_VIRTUALENV_ROOT="${PYENV_ROOT}/plugins/pyenv-virtualenv"
PYTHON_VERSION=2.7.9

if [ -d "${PYENV_ROOT}/.git" ]
then
  info "Updating pyenv"
  pushd "${PYENV_ROOT}"
  git pull --quiet
  pushd "${PYENV_VIRTUALENV_ROOT}"
  git pull --quiet
  popd
  popd
else
  info "Installing pyenv"
  git clone --depth=1 --quiet https://github.com/yyuu/pyenv.git "${PYENV_ROOT}"
  git clone --depth=1 --quiet https://github.com/yyuu/pyenv-virtualenv.git "${PYENV_VIRTUALENV_ROOT}"
fi

PATH="${PYENV_ROOT}/bin:${PATH}"

eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"

if [ ! "${PYTHON_VERSION}" = "system" ]
then
  pyenv install -s "${PYTHON_VERSION}"
fi

pyenv shell "${PYTHON_VERSION}"
pyenv versions

if pyenv virtualenvs | grep -q "${WERCKER_STEP_NAME}"
then
  info "Found existing virtualenv ${WERCKER_STEP_NAME}"
else 
  info "Creating virtualenv ${WERCKER_STEP_NAME}" 
  pyenv virtualenv "${WERCKER_STEP_NAME}"
fi

pyenv shell "${WERCKER_STEP_NAME}"

if [ "${WERCKER_STEP_NAME}" = "script" ]
then
  # This is test run during the step's build.
  REQS_DIR=${WERCKER_SOURCE_DIR}
else
  REQS_DIR=${WERCKER_STEP_ROOT}
fi

pip install -q -r "${REQS_DIR}/requirements.txt"

yolk -l

