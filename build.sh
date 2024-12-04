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

setup_cidoer() {
  local ref="${1:-main}" dir="${2:-.cidoer}" path
  local archive_url="https://github.com/i3ash/cidoer/archive/$ref.zip"
  printf '%s%s\n' 'Downloading:' "$archive_url"
  curl -fsSL "$archive_url" -o source.zip
  printf '%s%s\n' 'Extracting:' "$(pwd)"
  unzip -q source.zip -d "$(pwd)"
  rm source.zip
  rm -rf "$(pwd)"/"${dir:?}"
  mv "cidoer-$ref" "$dir"
  ls -lhAR "$dir"
  path="$(pwd)"/"$dir"
  source "$path"/cidoer.core.sh
  export CIDOER_DIR="$path"
  export CIDOER_CORE_FILE="$path/cidoer.core.sh"
  do_print_section FINISHED
  uname -a || print 'uname error'
  do_print_dash_pair 'CIDOER_OS_TYPE' "$(do_os_type)"
}

if [ ! -f '.cidoer/cidoer.core.sh' ];then
  setup_cidoer '' ''
fi
cleanup

source cidoer.sh
do_workflow_job prepare
do_workflow_job build

if [[ 'darwin' == $(uname -s | tr '[:upper:]' '[:lower:]') ]]; then
  do_workflow_job build_darwin_x64
  do_workflow_job build_darwin_universal
fi
