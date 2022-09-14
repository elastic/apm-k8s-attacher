// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLParse(t *testing.T) {
	p := "testdata/good.yaml"
	cfg, err := parseConfig(p)
	require.NoError(t, err)

	assert.Contains(t, cfg.Agents, "java")
	java := cfg.Agents["java"]
	assert.Equal(t, java.Image, "docker.com/elastic/agent-java:1.2.3")
	assert.Len(t, java.Environment, 5)
	assert.Contains(t, java.Environment, "ELASTIC_APM_SERVER_URLS")
	assert.Contains(t, java.Environment, "ELASTIC_APM_SERVICE_NAME")
	assert.Contains(t, java.Environment, "ELASTIC_APM_ENVIRONMENT")
	assert.Contains(t, java.Environment, "ELASTIC_APM_LOG_LEVEL")
	assert.Contains(t, java.Environment, "ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED")
}

func TestYAMLParseBad(t *testing.T) {
	p := "testdata/bad.yaml"
	cfg, err := parseConfig(p)
	require.Error(t, err)
	require.Nil(t, cfg)
	require.Equal(t, `custom agent "java" is missing 'image' value`, err.Error())

	p = "testdata/bad2.yaml"
	cfg, err = parseConfig(p)
	require.Error(t, err)
	require.Nil(t, cfg)
	require.Equal(t, `custom agent "java" is missing 'artifact' value`, err.Error())
}
