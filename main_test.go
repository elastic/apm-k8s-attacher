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
