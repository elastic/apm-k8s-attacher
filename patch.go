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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// patchOperation is an operation of a JSON patch, see
// https://tools.ietf.org/html/rfc6902.
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

const (
	volumeName = "elastic-apm-agent"
	mountPath  = "/elastic/apm/agent"
)

var (
	volumeMounts = corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
	agentVolume = corev1.Volume{
		Name:         volumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
	kubernetesEnvironmentVariables = map[string]string{
		"KUBERNETES_NODE_NAME": "spec.nodeName",
		"KUBERNETES_POD_NAME":  "metadata.name",
		"KUBERNETES_NAMESPACE": "metadata.namespace",
		"KUBERNETES_POD_UID":   "metadata.uid",
	}
)

func createPatch(config agentConfig, spec corev1.PodSpec) ([]patchOperation, error) {
	// Create patch operations
	var patches []patchOperation

	envVariables, err := generateEnvironmentVariables(config)
	if err != nil {
		return nil, err
	}

	// Add a volume mount to the pod
	patches = append(patches, createVolumePatch(spec.Volumes == nil))

	// Add an init container, that will fetch the agent Docker image and
	// extract the agent jar to the agent volume
	patch, err := createInitContainerPatch(config, spec.InitContainers == nil)
	if err != nil {
		return nil, err
	}
	patches = append(patches, patch)

	// Add agent env variables for each container at the pod, as well as
	// the volume mount
	containers := spec.Containers
	for index, container := range containers {
		// Filter out environment variables from the agent config to
		// not overwrite environment variables set directly on the
		// container.
		envVars := uniqueEnvironmentVariables(envVariables, container.Env)
		patches = append(patches, createVolumeMountsPatch(container.VolumeMounts == nil, index))
		patches = append(patches, createEnvVariablesPatches(envVars, container.Env == nil, index)...)
	}
	return patches, nil
}

func uniqueEnvironmentVariables(configEnvironmentVariables, containerEnvironmentVariables []corev1.EnvVar) []corev1.EnvVar {
	if len(containerEnvironmentVariables) == 0 {
		return configEnvironmentVariables
	}
	unique := make([]corev1.EnvVar, 0, len(configEnvironmentVariables))
	containerKeys := make(map[string]struct{}, len(containerEnvironmentVariables))
	for _, v := range containerEnvironmentVariables {
		containerKeys[v.Name] = struct{}{}
	}
	for _, v := range configEnvironmentVariables {
		if _, ok := containerKeys[v.Name]; ok {
			continue
		}
		unique = append(unique, v)
	}
	return unique
}

func generateEnvironmentVariables(config agentConfig) ([]corev1.EnvVar, error) {
	optional := true
	vars := []corev1.EnvVar{{
		Name: "ELASTIC_APM_SECRET_TOKEN",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "apm-agent-auth"},
				Key:                  "secret_token",
				Optional:             &optional,
			},
		},
	}, {
		Name: "ELASTIC_APM_API_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "apm-agent-auth"},
				Key:                  "api_key",
				Optional:             &optional,
			},
		},
	}}
	if _, ok := config.Environment["ELASTIC_APM_SERVER_URL"]; !ok {
		// No apm-server url present, inject the local node address.
		vars = append(vars, corev1.EnvVar{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		}, corev1.EnvVar{
			Name: "ELASTIC_APM_SERVER_URL", Value: "http://$(HOST_IP):8200",
		})
	}
	if _, ok := config.Environment["ELASTIC_APM_ACTIVATION_METHOD"]; !ok {
		vars = append(vars, corev1.EnvVar{
			Name: "ELASTIC_APM_ACTIVATION_METHOD", Value: "K8S_ATTACH",
		})
	}
	for k, v := range kubernetesEnvironmentVariables {
		vars = append(vars, corev1.EnvVar{
			Name: k,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: v,
				},
			},
		})
	}
	for name, value := range config.Environment {
		vars = append(vars, corev1.EnvVar{Name: name, Value: value})
	}
	for _, ev := range vars {
		if errs := validation.IsEnvVarName(ev.Name); len(errs) != 0 {
			return nil, fmt.Errorf("failed to validate environment variable %q: %v", ev.Name, errs)
		}
	}
	return vars, nil
}

func createVolumePatch(createArray bool) patchOperation {
	if createArray {
		return patchOperation{
			Op:    "add",
			Path:  "/spec/volumes",
			Value: []corev1.Volume{agentVolume},
		}
	}
	return patchOperation{
		Op:    "add",
		Path:  "/spec/volumes/-",
		Value: agentVolume,
	}
}

func createInitContainerPatch(config agentConfig, createArray bool) (patchOperation, error) {
	bp := filepath.Base(config.Image)
	name := strings.Split(bp, ":")
	allowPrivilegeEscalation := false
	agentInitContainer := corev1.Container{
		Name:         name[0],
		Image:        config.Image,
		VolumeMounts: []corev1.VolumeMount{volumeMounts},
		// TODO: should this be a default, and then users can modify it
		// *if needed*?
		Command: []string{"cp", "-v", "-r", config.ArtifactPath, mountPath},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	}

	if errs := validation.IsDNS1123Label(agentInitContainer.Name); len(errs) != 0 {
		return patchOperation{}, fmt.Errorf("failed to extract container name from image (%s): init container name (%s) is not a valid DNS_LABEL: %v",
			config.Image, agentInitContainer.Name, errs,
		)
	}

	if createArray {
		return patchOperation{
			Op:    "add",
			Path:  "/spec/initContainers",
			Value: []corev1.Container{agentInitContainer},
		}, nil
	}
	return patchOperation{
		Op:    "add",
		Path:  "/spec/initContainers/-",
		Value: agentInitContainer,
	}, nil
}

// If the env variable array does not already exist, this method will return a
// single patch operation for the addition of the entire list, otherwise it
// would return a list of patches for each env variable
func createEnvVariablesPatches(envVariables []corev1.EnvVar, createArray bool, index int) []patchOperation {
	containerIndex := strconv.Itoa(index)
	envVariablesPath := "/spec/containers/" + containerIndex + "/env"
	if createArray {
		return []patchOperation{{
			Op:    "add",
			Path:  envVariablesPath,
			Value: envVariables,
		}}
	}
	patches := make([]patchOperation, len(envVariables))
	envVariablesPath += "/-"
	for i, variable := range envVariables {
		patches[i] = patchOperation{
			Op:    "add",
			Path:  envVariablesPath,
			Value: variable,
		}
	}
	return patches
}

func createVolumeMountsPatch(createArray bool, index int) patchOperation {
	containerIndex := strconv.Itoa(index)
	volumeMountsPath := "/spec/containers/" + containerIndex + "/volumeMounts"
	if createArray {
		return patchOperation{
			Op:    "add",
			Path:  volumeMountsPath,
			Value: []corev1.VolumeMount{volumeMounts},
		}
	}
	return patchOperation{
		Op:    "add",
		Path:  volumeMountsPath + "/-",
		Value: volumeMounts,
	}
}
