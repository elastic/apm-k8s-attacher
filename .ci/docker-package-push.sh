#!/usr/bin/env bash
#
# Build the docker image for the mutating webhook and push it to the given
# docker registry.
#
# Arguments:
# - REPO: the docker repo name
# - NAME: the docker image name
# - TAG: the docker tag version
# - SHORT_GIT_SHA: whether to short the git sha, aka the given TAG argument.

set -euo pipefail

export REPO=${1:?docker repo not set}
export NAME=${2:?docker image name not set}
export TAG=${3:?docker tag not set}
export SHORT_GIT_SHA=${4:-false}

# Use the short git SHA
if [ "$SHORT_GIT_SHA" == "true" ] ; then
	TAG=${TAG:0:7}
fi

fqn="${REPO}/${NAME}:${TAG}"
latest="${REPO}/${NAME}:latest"

echo "INFO: Build docker image ${fqn}"
make .webhook

echo "INFO: Push docker image ${fqn}"
docker push "${fqn}"
