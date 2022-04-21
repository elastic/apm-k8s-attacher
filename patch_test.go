package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestConfig(t *testing.T) {
	config := agentConfig{
		container:   "stuartnelson3/agent-container",
		environment: map[string]string{"SANITY_CHECK": "true"},
	}
	vars := generateEnvironmentVariables(config)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 2)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "SANITY_CHECK", environmentVariables[1].Name)
}
