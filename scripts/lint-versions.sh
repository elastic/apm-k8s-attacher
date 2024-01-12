#!/usr/bin/env bash

# Copyright 2022 Elasticsearch BV
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Doing a release involves updating the version in two places. This script
# ensures those versions match.

if [ "$TRACE" != "" ]; then
    export PS4='${BASH_SOURCE}:${LINENO}: ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
    set -o xtrace
fi
set -o errexit
set -o pipefail

TOP=$(unset CDPATH; cd $(dirname $0)/../; pwd)

function fatal {
    echo "$(basename $0): error: $*"
    exit 1
}

CHART_PATH=charts/apm-attacher/Chart.yaml
CHART_VER=$(grep "^version:" "$TOP/$CHART_PATH" | cut -d'"' -f2)
VALUES_PATH=charts/apm-attacher/values.yaml
VALUES_VER=$(grep "tag:" "$TOP/$VALUES_PATH" | cut -d'"' -f2)

if [[ "${VALUES_VER:0:1}" != "v" ]]; then
    fatal "tag value in $VALUES_PATH, '$VALUES_VER', does not start with a 'v'"
fi
if [[ "v$CHART_VER" != "$VALUES_VER" ]]; then
    fatal "version in $CHART_PATH, '$CHART_VER', does not match tag in $VALUES_PATH, '$VALUES_VER'"
fi
