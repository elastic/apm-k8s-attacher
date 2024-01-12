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
	vars, errs := generateEnvironmentVariables(config)
	assert.Nil(t, errs)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 9)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "ELASTIC_APM_API_KEY", environmentVariables[1].Name)
	assert.Equal(t, "ELASTIC_APM_ACTIVATION_METHOD", environmentVariables[2].Name)

	m := map[string]struct{}{
		"KUBERNETES_NODE_NAME": {},
		"KUBERNETES_POD_NAME":  {},
		"KUBERNETES_NAMESPACE": {},
		"KUBERNETES_POD_UID":   {},
	}
	for _, envVar := range environmentVariables {
		delete(m, envVar.Name)
	}
	assert.Len(t, m, 0)
	customVars := environmentVariables[7:]
	sort.Slice(customVars, func(i, j int) bool {
		return customVars[i].Name < customVars[j].Name
	})
	assert.Equal(t, "ELASTIC_APM_SERVER_URL", customVars[0].Name)
	assert.Equal(t, "SANITY_CHECK", customVars[1].Name)
}

func TestAddAPMServerURLIfNotPresent(t *testing.T) {
	config := agentConfig{Environment: map[string]string{}}
	vars, errs := generateEnvironmentVariables(config)
	assert.Nil(t, errs)
	patches := createEnvVariablesPatches(vars, true, 0)
	assert.Len(t, patches, 1)
	environmentVariables := patches[0].Value.([]corev1.EnvVar)
	assert.Len(t, environmentVariables, 9)
	assert.Equal(t, "ELASTIC_APM_SECRET_TOKEN", environmentVariables[0].Name)
	assert.Equal(t, "ELASTIC_APM_API_KEY", environmentVariables[1].Name)
	assert.Equal(t, "HOST_IP", environmentVariables[2].Name)
	assert.Equal(t, "ELASTIC_APM_SERVER_URL", environmentVariables[3].Name)
	assert.Equal(t, "http://$(HOST_IP):8200", environmentVariables[3].Value)
	assert.Equal(t, "ELASTIC_APM_ACTIVATION_METHOD", environmentVariables[4].Name)
	assert.Equal(t, "K8S_ATTACH", environmentVariables[4].Value)
}

func TestInitContainerPatch(t *testing.T) {
	config := agentConfig{
		Image:        "stuartnelson3/agent-image:1.2.3",
		ArtifactPath: "/tmp/artifact.jar",
	}
	patch, err := createInitContainerPatch(config, true)
	assert.NoError(t, err)
	initContainers := patch.Value.([]corev1.Container)
	assert.Len(t, initContainers, 1)
	ic := initContainers[0]
	assert.Equal(t, ic.Name, "agent-image")
	assert.Equal(t, ic.Image, config.Image)
	assert.Equal(t, ic.Command, []string{"cp", "-v", "-r", config.ArtifactPath, mountPath})
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
