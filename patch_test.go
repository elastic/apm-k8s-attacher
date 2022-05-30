package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVarPatch(t *testing.T) {
	config := agentConfig{
		Environment: map[string]string{
			"SANITY_CHECK":           "true",
			"ELASTIC_APM_SERVER_URL": "abc",
		},
	}
	vars := generateEnvironmentVariables(config)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 8)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "ELASTIC_APM_API_KEY", environmentVariables[1].Name)

	m := map[string]struct{}{
		"KUBERNETES_NODE_NAME": {},
		"KUBERNETES_POD_NAME":  {},
		"KUBERNETES_NAMESPACE": {},
		"KUBERNETES_POD_UID":   {},
	}
	for _, envVar := range environmentVariables {
		if _, ok := m[envVar.Name]; ok {
			delete(m, envVar.Name)
		}
	}
	assert.Len(t, m, 0)
	customVars := environmentVariables[6:]
	sort.Slice(customVars, func(i, j int) bool {
		return customVars[i].Name < customVars[j].Name
	})
	assert.Equal(t, "ELASTIC_APM_SERVER_URL", customVars[0].Name)
	assert.Equal(t, "SANITY_CHECK", customVars[1].Name)
}

func TestAddAPMServerURLIfNotPresent(t *testing.T) {
	config := agentConfig{Environment: map[string]string{}}
	vars := generateEnvironmentVariables(config)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 8)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "ELASTIC_APM_API_KEY", environmentVariables[1].Name)
	assert.Equal(t, "HOST_IP", environmentVariables[2].Name)
	assert.Equal(t, "ELASTIC_APM_SERVER_URL", environmentVariables[3].Name)
	assert.Equal(t, "http://$(HOST_IP):8200", environmentVariables[3].Value)
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

func TestUniqueEnvironmentVariables(t *testing.T) {
	webhook := []corev1.EnvVar{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "1"},
		{Name: "c", Value: "1"},
		{Name: "f", Value: "1"},
	}
	container := []corev1.EnvVar{
		{Name: "b", Value: "0"},
		{Name: "c", Value: "0"},
		{Name: "d", Value: "0"},
	}
	expected := []corev1.EnvVar{
		{Name: "a", Value: "1"},
		{Name: "f", Value: "1"},
	}
	unique := uniqueEnvironmentVariables(webhook, container)

	assert.Equal(t, expected, unique)
}
