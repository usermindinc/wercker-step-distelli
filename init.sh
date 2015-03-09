#!/bin/bash

function _message() {
    msg=$1
    color=$2
    echo "${color}${msg}${RESET_COLOR}"
}

function success() {
    _message "${1}" $SUCCESS_COLOR
}

function info() {
    _message "${1}" $INFO_COLOR
}

function debug() {
    _message "${1}" $DEBUG_COLOR
}

function warn() {
    _message "${1}" $WARN_COLOR
}

function error() {
    _message "error: ${1}" $ERROR_COLOR
}

function fail() {
    _message "failed: ${1}" $ERROR_COLOR
    echo "${1}" > "$WERCKER_REPORT_MESSAGE_FILE"
    exit 1
}


function setMessage() {
  echo "${1}" > "$WERCKER_REPORT_MESSAGE_FILE"
}



# Make sure we fail on all errors
set -e

