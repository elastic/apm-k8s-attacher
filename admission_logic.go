package main

import (
	"fmt"
	"log"
	"strconv"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	traceLabel = `elastic_trace`
)

var (
	volumeMounts = corev1.VolumeMount{
		Name:      "elastic-apm-agent",
		MountPath: "/elastic/apm/agent",
	}

	envVariables = []corev1.EnvVar{
		{Name: "ELASTIC_APM_SERVER_URLS", Value: "http://34.78.173.219:8200"},
		{Name: "ELASTIC_APM_SERVICE_NAME", Value: "petclinic"},
		{Name: "ELASTIC_APM_ENVIRONMENT", Value: "test"},
		{Name: "ELASTIC_APM_LOG_LEVEL", Value: "debug"},
		{Name: "ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED", Value: "true"},
		{Name: "ELASTIC_APM_SECRET_TOKEN", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "apm-server-apm-token"},
				Key:                  "secret-token",
			},
		}},
		{Name: "JAVA_TOOL_OPTIONS", Value: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"},
	}
)

func createVolumePatch(createArray bool) patchOperation {
	agentVolume := corev1.Volume{
		Name:         "elastic-apm-agent",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
	var patch patchOperation
	if createArray {
		patch = patchOperation{
			Op:    "add",
			Path:  "/spec/volumes",
			Value: []corev1.Volume{agentVolume},
		}
	} else {
		patch = patchOperation{
			Op:    "add",
			Path:  "/spec/volumes/-",
			Value: agentVolume,
		}
	}
	return patch
}

func createInitContainerPatch(createArray bool) patchOperation {
	agentInitContainer := corev1.Container{
		Name:         "elastic-java-agent",
		Image:        "docker.elastic.co/observability/apm-agent-java:1.23.0",
		VolumeMounts: []corev1.VolumeMount{volumeMounts},
		Command:      []string{"cp", "-v", "/usr/agent/elastic-apm-agent.jar", "/elastic/apm/agent"},
	}
	var patch patchOperation
	if createArray {
		patch = patchOperation{
			Op:    "add",
			Path:  "/spec/initContainers",
			Value: []corev1.Container{agentInitContainer},
		}
	} else {
		patch = patchOperation{
			Op:    "add",
			Path:  "/spec/initContainers/-",
			Value: agentInitContainer,
		}
	}
	return patch
}

// If the evn variable array does not already exist, this method will return a single patch operation for the addition of the entire list,
// otherwise it would return a list of patches for each env variable
func createEnvVariablesPatches(createArray bool, index int) []patchOperation {
	containerIndex := strconv.Itoa(index)
	envVariablesPath := "/spec/containers/" + containerIndex + "/env"
	var patches []patchOperation
	if createArray {
		patches = []patchOperation{{
			Op:    "add",
			Path:  envVariablesPath,
			Value: envVariables,
		}}
	} else {
		envVariablesPath += "/-"
		patches = []patchOperation{}
		for _, variable := range envVariables {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  envVariablesPath,
				Value: variable,
			})
		}
	}
	return patches
}

func createVolumeMountsPatch(createArray bool, index int) patchOperation {
	containerIndex := strconv.Itoa(index)
	volumeMountsPath := "/spec/containers/" + containerIndex + "/volumeMounts"
	var patch patchOperation
	if createArray {
		patch = patchOperation{
			Op:    "add",
			Path:  volumeMountsPath,
			Value: []corev1.VolumeMount{volumeMounts},
		}
	} else {
		patch = patchOperation{
			Op:    "add",
			Path:  volumeMountsPath + "/-",
			Value: volumeMounts,
		}
	}
	return patch
}

var podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}

// Applying the proper settings in order to trace all containers of pods that has the tracing label
func applyTraceSettings(req *admissionv1.AdmissionRequest) ([]patchOperation, error) {
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}

	// Parse the Pod object.
	raw := req.Object.Raw
	pod := corev1.Pod{}
	// todo: use the returned pod may be safer than using the provided reference, as per the documentation of Decode
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	fmt.Fprintf(log.Writer(), "(version #5) pod configuration: %+v\n", pod)

	labels := pod.Labels
	if labels == nil {
		return nil, nil
	}

	// We apply our patches only if the tracing label is set on this pod
	traceValue, ok := labels[traceLabel]
	if !ok {
		return nil, nil
	} else {
		parseBool, err := strconv.ParseBool(traceValue)
		if err != nil {
			log.Printf("the `%s` label should be set with a boolean value. `%s` is illegal", traceLabel, traceValue)
		}
		if !parseBool {
			return nil, nil
		}
	}

	spec := pod.Spec

	// Create patch operations
	var patches []patchOperation

	// Add a volume mount to the pod
	patches = append(patches, createVolumePatch(spec.Volumes == nil))

	// Add an init container, that will fetch the agent Docker image and extract the agent jar to the agent volume
	patches = append(patches, createInitContainerPatch(spec.InitContainers == nil))

	// Add agent env variables for each container at the pod, as well as the volume mount
	containers := spec.Containers
	for index, container := range containers {
		patches = append(patches, createVolumeMountsPatch(container.VolumeMounts == nil, index))
		patches = append(patches, createEnvVariablesPatches(container.Env == nil, index)...)
	}

	return patches, nil
}
