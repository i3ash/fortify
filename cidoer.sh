#!/usr/bin/env bash
# shellcheck disable=SC2317
set -eou pipefail

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

if [ -f '.cidoer/cidoer.core.sh' ];then
  source .cidoer/cidoer.core.sh
else setup_cidoer '1.0';  fi

declare -rx ARTIFACT_CMD='fortify'

define_custom_test() {
  test_custom_do() {
    go test -v ./...
  }
}

define_custom_prepare() {
  prepare_custom_do() {
    local tag
    tag=$(custom_version_tag)
    export ARTIFACT_TAG="$tag"
    do_print_dash_pair 'ARTIFACT_TAG' "$ARTIFACT_TAG"
    sed -e "s|#VERSION|${tag}|g" < "cmd/version.go-e" > "cmd/version.go"
    check_go
  }
  custom_version_tag() {
    local tag count hash
    tag=$(do_git_version_tag)
    count=$(do_git_count_commits_since "$tag")
    hash=$(do_git_short_commit_hash)
    if [ "$count" -gt 0 ]; then
      printf '%s%s%s' "$tag" ".$count" "-$hash"
    else printf '%s' "$tag"; fi
  }
  check_go() {
    if ! command -v go &>/dev/null; then
      echo "Error: Go language environment not found." >&2
      exit 1
    fi
    printf "Check: " >&1
    which go >&1
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
  do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
}

define_custom_build_linux_x64() {
  build_linux_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-x86_64"
    build_artifact linux amd64
  }
}

define_custom_build_linux_aarch64() {
  build_linux_aarch64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-aarch64"
    build_artifact linux arm64
  }
}

define_custom_build_linux_riscv64() {
  build_linux_riscv64_custom_do() {
    build_artifact linux riscv64
  }
}

define_custom_build_linux_mips64le() {
  build_linux_mips64le_custom_do() {
    build_artifact linux mips64le
  }
}

define_custom_build_windows_x64() {
  build_windows_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-x86_64.exe"
    build_artifact windows amd64
  }
}

define_custom_build_windows_aarch64() {
  build_windows_aarch64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-aarch64.exe"
    build_artifact windows arm64
  }
}

define_custom_build_darwin_arm64() {
  build_darwin_arm64_custom_do() {
    build_artifact darwin arm64
  }
}

define_custom_build_darwin_x64() {
  build_darwin_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-darwin-x86_64"
    build_artifact darwin amd64
  }
}

define_custom_build_darwin_universal() {
  build_darwin_universal_custom_do() {
    export ARTIFACT_FILENAME="${ARTIFACT_CMD}-darwin-universal"
    do_print_dash_pair 'ARTIFACT_FILENAME' "$ARTIFACT_FILENAME"
    local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
    local cmd1="${ARTIFACT_DARWIN_X64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-x86_64")}"
    local cmd2="${ARTIFACT_DARWIN_A64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-arm64")}"
    lipo -create -output "$out" "$cmd1" "$cmd2"
    file "$out"
    do_print_dash_pair 'ARTIFACT_OUT' "$out"
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
  }
}


define_custom_docker_debian() {
  docker_debian_custom_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:debian")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/amd64,linux/arm64/v8,linux/arm/v7,linux/arm/v5,linux/386,linux/ppc64le,linux/s390x,linux/mips64le \
      --target debian --push .
  }
}

define_custom_docker_alpine() {
  docker_alpine_custom_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:alpine")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/amd64,linux/arm64/v8,linux/arm/v7,linux/arm/v6,linux/386,linux/ppc64le,linux/s390x,linux/riscv64 \
      --target alpine --push .
  }
}

define_custom_docker_minimal() {
  docker_minimal_custom_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:latest")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/amd64,linux/arm64/v8,linux/arm/v7,linux/arm/v6,linux/386,linux/ppc64le,linux/s390x,linux/riscv64 \
      --target minimal --push .
  }
}
