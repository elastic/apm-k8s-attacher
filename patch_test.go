package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVarPatch(t *testing.T) {
	config := agentConfig{
		Environment: map[string]string{"SANITY_CHECK": "true"},
	}
	vars := generateEnvironmentVariables(config)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 2)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "SANITY_CHECK", environmentVariables[1].Name)
}

func TestInitContainerPatch(t *testing.T) {
	config := agentConfig{
		Image:        "stuartnelson3/agent-image:1.2.3",
		ArtifactPath: "/tmp/artifact.jar",
	}
	patch := createInitContainerPatch(config, true)
	initContainers := patch.Value.([]corev1.Container)
	assert.Len(t, initContainers, 1)
	ic := initContainers[0]
	assert.Equal(t, ic.Name, "agent-image")
	assert.Equal(t, ic.Image, config.Image)
	assert.Equal(t, ic.Command, []string{"cp", "-v", config.ArtifactPath, mountPath})
	assert.NotEmpty(t, ic.VolumeMounts)
}
