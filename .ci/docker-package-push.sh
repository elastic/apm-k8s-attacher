#!/usr/bin/env bash
#
# Build the docker image for the mutating webhook and push it to the given
# docker registry.
#
# Arguments:
# - REPO: the docker repo name
# - NAME: the docker image name
# - TAG: the docker tag version

set -euo pipefail

export REPO=${1:?docker repo not set}
export NAME=${2:?docker image name not set}
export TAG=${3:?docker tag not set}

# Use the short git SHA
TAG=${TAG:0:7}

fqn="${REPO}/${NAME}:${TAG}"
latest="${REPO}/${NAME}:latest"

echo "INFO: Build docker image ${fqn}"
make .webhook

echo "INFO: Push docker image ${fqn}"
docker push "${fqn}"

echo "INFO: Push docker image ${latest}"
docker tag "${fqn}" "${latest}"
docker push "${latest}"
