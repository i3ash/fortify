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

if [[ 'darwin' == $(do_os_type) ]]; then
  do_workflow_job build_darwin_arm64
  do_workflow_job build_darwin_x64
  do_workflow_job build_darwin_universal
  ln -s "${OUT_DIR}/${ARTIFACT_CMD:?}-darwin-$(uname -m)" "${OUT_DIR}/${ARTIFACT_CMD:?}"
elif [[ 'linux' == $(do_os_type) ]]; then
  do_workflow_job build_linux_x64
  do_workflow_job build_linux_aarch64
  do_workflow_job build_linux_riscv64
  do_workflow_job build_linux_mips64le
  ln -s "${OUT_DIR}/${ARTIFACT_CMD:?}-linux-$(uname -m)" "${OUT_DIR}/${ARTIFACT_CMD:?}"
elif [[ 'windows' == $(do_os_type) ]]; then
  do_workflow_job build_windows_x64
  do_workflow_job build_windows_aarch64
  ln -s "${OUT_DIR}/${ARTIFACT_CMD:?}-windows-$(uname -m)" "${OUT_DIR}/${ARTIFACT_CMD:?}"
fi

ls -lhA "${OUT_DIR}"
