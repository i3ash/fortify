#!/usr/bin/env bash

[ -f cidoer.sh ] || { [ -f ../cidoer.sh ] && cd ..; } || { echo 'No cidoer.sh' >&2 && exit 127; }
source cidoer.sh || exit $?

LATEST_TAG="${LATEST_TAG:-false}"
do_workflow_job prepare

declare -r TAG_PREFIX="${ARTIFACT_TAG:?}"

ARTIFACT_TAG="${TAG_PREFIX}-debian"
do_workflow_job docker_debian

ARTIFACT_TAG="${TAG_PREFIX}-alpine"
do_workflow_job docker_alpine

ARTIFACT_TAG="${TAG_PREFIX}-distroless"
do_workflow_job docker_distroless

ARTIFACT_TAG="${TAG_PREFIX}-nonroot"
do_workflow_job docker_distroless_nonroot

ARTIFACT_TAG="${TAG_PREFIX}"
do_workflow_job docker_minimal
