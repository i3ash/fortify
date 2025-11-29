#!/usr/bin/env bash
# shellcheck disable=SC2317
set -eu -o pipefail

declare -rx ARTIFACT_CMD='fortify'

define_test() {
  test_do() {
    go test -v ./...
  }
}

build_artifact() {
  local os="${1:?Require os type}"
  local arch="${2:?Require architecture}"
  export ARTIFACT_FILENAME="${ARTIFACT_FILENAME:-$(printf '%s' "$ARTIFACT_CMD-$os-$arch")}"
  do_print_dash_pair 'ARTIFACT_FILENAME' "${ARTIFACT_FILENAME}"
  local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
  GOARCH=$arch GOOS=$os CGO_ENABLED=0 go build -a -installsuffix nocgo -ldflags '-s -w' -o "$out"
  file "$out"
  do_print_dash_pair 'ARTIFACT_OUT' "$out"
  if [ 'darwin' = "$(do_os_type)" ] || [[ "${ARTIFACT_FILENAME:-}" == *"$(do_host_type)"* ]]; then
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
  fi
}

define_build_linux_x64() {
  build_linux_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-x86_64"
    build_artifact linux amd64
  }
}

define_build_linux_arm64() {
  build_linux_arm64_do() {
    build_artifact linux arm64
  }
}

define_build_linux_riscv64() {
  build_linux_riscv64_do() {
    build_artifact linux riscv64
  }
}

define_build_linux_mips64le() {
  build_linux_mips64le_do() {
    build_artifact linux mips64le
  }
}

define_build_windows_x64() {
  build_windows_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-x86_64.exe"
    build_artifact windows amd64
  }
}

define_build_windows_arm64() {
  build_windows_arm64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-arm64.exe"
    build_artifact windows arm64
  }
}

define_build_darwin_arm64() {
  build_darwin_arm64_do() {
    build_artifact darwin arm64
  }
}

define_build_darwin_x64() {
  build_darwin_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-darwin-x86_64"
    build_artifact darwin amd64
  }
}

define_build_darwin_universal() {
  build_darwin_universal_do() {
    do_print_dash_pair "${FUNCNAME[0]}"
    export ARTIFACT_FILENAME="${ARTIFACT_CMD}-darwin-universal"
    do_print_dash_pair 'ARTIFACT_FILENAME' "$ARTIFACT_FILENAME"
    local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
    local cmd1="${ARTIFACT_DARWIN_X64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-x86_64")}"
    local cmd2="${ARTIFACT_DARWIN_A64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-arm64")}"
    lipo -create -output "$out" "$cmd1" "$cmd2"
    file "$out"
    chmod +x "$out"
    do_print_dash_pair 'ARTIFACT_OUT' "$out"
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
    $out version -d
    do_print_trace "$(do_stack_trace)" done!
  }
}

define_cidoer_core() {
  [ -f cidoer.print.sh ] && source cidoer.print.sh
  declare -F '_print_defined' >/dev/null || { declare -F 'define_cidoer_print' >/dev/null && define_cidoer_print; }
  declare -F '_core_defined' >/dev/null && return 0
  _core_defined() { :; }
  CIDOER_DEBUG='no'
  CIDOER_OS_TYPE=''
  CIDOER_HOST_TYPE=''
  do_workflow_job() {
    local -r job_type=$(do_trim "${1:-}")
    [[ "${job_type:-}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] || {
      do_print_warn "$(do_stack_trace)" $'$1 (job_type) is invalid:' "'${1:-}'" >&2
      return 1
    }
    local -a steps=()
    local arg step
    for arg in "${@:2}"; do
      step=$(do_trim "$arg")
      [[ "${step:-}" =~ ^[a-zA-Z0-9_]*$ ]] || {
        do_print_warn "$(do_stack_trace)" "step name of '${job_type:-}' is invalid:" "'${step:-}'" >&2
        return 1
      }
      steps+=("$step")
    done
    [ ${#steps[@]} -le 0 ] && steps=('do')
    local -r lower=$(do_convert_to_lower "$job_type")
    for step in "${steps[@]}"; do
      declare -F "${lower}_${step}" >/dev/null && local -r defined=1 && break
    done
    [ "${defined:-0}" -eq 1 ] || {
      do_func_invoke "define_${lower}" "${steps[@]}" || return $?
    }
    for step in "${steps[@]}"; do
      do_func_invoke "${lower}_${step}" || return $?
    done
  }
  do_func_invoke() {
    local -r func_name="${1:-}"
    local -r func_finally="${func_name}_finally"
    [ -z "$func_name" ] && {
      do_print_warn "$(do_stack_trace)" $'$1 (func_name) is required' >&2
      return 0
    }
    declare -F "$func_name" >/dev/null || {
      do_print_trace "$(do_stack_trace)" "$func_name is an absent function" >&2
      return 0
    }
    "${@}" || local -r status=$?
    declare -F "$func_finally" >/dev/null && {
      [ "${status:-0}" -eq 0 ] || do_print_info "$(do_stack_trace)" "$func_name failed with exit code $status" >&2
      "$func_finally" "${status:-0}" || local -r code=$?
      [ "${code:-0}" -eq 0 ] || do_print_warn "$(do_stack_trace)" "$func_finally failed with exit code $code" >&2
      return "${code:-0}"
    }
    [ "${status:-0}" -eq 0 ] || do_print_warn "$(do_stack_trace)" "$func_name failed with exit code $status" >&2
    return "${status:-0}"
  }
  do_trim() {
    local var="${1:-}"
    var="${var#"${var%%[![:space:]]*}"}"
    var="${var%"${var##*[![:space:]]}"}"
    printf '%s' "$var"
  }
  do_convert_to_lower() {
    local input="${1:-}"
    do_check_bash_4 && printf '%s\n' "${input,,}" && return 0
    printf '%s\n' "$input" | sed 'y/ABCDEFGHIJKLMNOPQRSTUVWXYZ/abcdefghijklmnopqrstuvwxyz/'
  }
  do_os_type() {
    [ -n "${CIDOER_OS_TYPE:-}" ] && {
      printf '%s\n' "${CIDOER_OS_TYPE:-}"
      return 0
    }
    if [ -z "${OSTYPE:-}" ]; then
      command -v uname >/dev/null 2>&1 && local -r os="$(uname -s)"
    else local -r os="${OSTYPE:-}"; fi
    local -r type=$(do_convert_to_lower "$os")
    case "${type:-}" in
    linux*) CIDOER_OS_TYPE='linux' ;;
    darwin*) CIDOER_OS_TYPE='darwin' ;;
    cygwin* | msys* | mingw* | windows*) CIDOER_OS_TYPE='windows' ;;
    *) CIDOER_OS_TYPE='unknown' ;;
    esac
    printf '%s\n' "${CIDOER_OS_TYPE:-}"
  }
  do_host_type() {
    [ -n "${CIDOER_HOST_TYPE:-}" ] && {
      printf '%s\n' "$CIDOER_HOST_TYPE"
      return 0
    }
    if [ -z "${HOSTTYPE:-}" ]; then
      command -v uname >/dev/null 2>&1 && local -r host="$(uname -m)"
    else local -r host="${HOSTTYPE:-}"; fi
    local -r type=$(do_convert_to_lower "$host")
    case "$type" in
    x86_64 | amd64 | x64) CIDOER_HOST_TYPE='x86_64' ;;
    i*86 | x86) CIDOER_HOST_TYPE='x86' ;;
    arm64 | aarch64) CIDOER_HOST_TYPE='arm64' ;;
    armv5* | armv6* | armv7* | aarch32) CIDOER_HOST_TYPE='arm' ;;
    armv8*) CIDOER_HOST_TYPE="$type" ;;
    ppc | powerpc) CIDOER_HOST_TYPE='ppc' ;;
    ppc64 | ppc64le) CIDOER_HOST_TYPE="$type" ;;
    mips | mips64 | mipsle | mips64le | s390x | riscv64) CIDOER_HOST_TYPE="$type" ;;
    *) CIDOER_HOST_TYPE='unknown' ;;
    esac
    printf '%s\n' "$CIDOER_HOST_TYPE"
  }
  do_check_bash_4() {
    [ -z "${BASH_VERSION:-}" ] && return 1
    [ "${BASH_VERSINFO[0]}" -lt 4 ] && return 1
    return 0
  }
  do_print_fix() {
    declare -F 'do_tint' >/dev/null || do_tint() { printf '%s\n' "- ${*:2}"; }
    declare -F 'do_print_with_color' >/dev/null || do_print_with_color() { return 1; }
    declare -F 'do_print_trace' >/dev/null || do_print_trace() { printf '%s\n' "- $*"; }
    declare -F 'do_print_info' >/dev/null || do_print_info() { printf '%s\n' "= $*"; }
    declare -F 'do_print_warn' >/dev/null || do_print_warn() { printf '%s\n' "? $*"; }
    declare -F 'do_print_error' >/dev/null || do_print_error() { printf '%s\n' "! $*"; }
    declare -F 'do_print_section' >/dev/null || do_print_section() { printf '%s\n' "== $*"; }
    declare -F 'do_print_dash_pair' >/dev/null || do_print_dash_pair() { printf '%s\n' "-- $*"; }
    declare -F 'do_print_code_lines' >/dev/null || do_print_code_lines() { printf '%s\n' "$*"; }
    declare -F 'do_print_code_bash' >/dev/null || do_print_code_bash() { do_print_code_lines "$@"; }
    declare -F 'do_print_code_bash_fn' >/dev/null || do_print_code_bash_fn() {
      do_print_code_bash "$(declare -f "$@")"
    }
    declare -F 'do_print_code_bash_debug' >/dev/null || do_print_code_bash_debug() {
      [ "${CIDOER_DEBUG:-no}" != "yes" ] && return 0
      do_print_code_bash "$@" >&2
    }
    declare -F 'do_print_debug' >/dev/null || do_print_debug() {
      [ "${CIDOER_DEBUG:-no}" != "yes" ] && return 0
      do_print_code_lines "$@" >&2
    }
    declare -F 'do_stack_trace' >/dev/null || do_stack_trace() {
      # shellcheck disable=SC2319
      local -ir status=$?
      local -i idx
      local -a filtered_fns=()
      for ((idx = ${#FUNCNAME[@]} - 2; idx > 0; idx--)); do
        [ 'do_func_invoke' != "${FUNCNAME[$idx]}" ] && filtered_fns+=("${FUNCNAME[$idx]}")
      done
      if [ ${#filtered_fns[@]} -gt 0 ]; then
        printf '%s --> %s\n' "${USER:-$(id -un)}@${HOSTNAME:-$(hostname)}" "${filtered_fns[*]}"
      else printf '%s -->\n' "${USER:-$(id -un)}@${HOSTNAME:-$(hostname)}"; fi
      return "$status"
    }
  }
  declare -F 'do_stack_trace' >/dev/null || do_print_fix
}
define_cidoer_core
