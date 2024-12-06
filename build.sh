#!/usr/bin/env bash

set -euo pipefail
declare SELF
SELF=$(readlink -f "$0")
declare -r SELF_DIR=${SELF%/*}
declare -r OUT_DIR=${SELF_DIR:?}/build

cleanup() {
  mkdir -p "${OUT_DIR}"
  if [ -d "${OUT_DIR}" ]; then
    rm -rf "${OUT_DIR:?}"/*
  fi
  mkdir -p "${OUT_DIR}/sss"
  mkdir -p "${OUT_DIR}/rsa"
}
cleanup

source cidoer.sh
do_workflow_job prepare

if [[ 'darwin' == $(uname -s | tr '[:upper:]' '[:lower:]') ]]; then
  do_workflow_job build_darwin_arm64
  do_workflow_job build_darwin_x64
  do_workflow_job build_darwin_universal
fi

ls -lhA "${OUT_DIR}"
