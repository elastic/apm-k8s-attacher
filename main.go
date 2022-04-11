package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func main() {
	var (
		runtimeScheme = runtime.NewScheme()
		codecs        = serializer.NewCodecFactory(runtimeScheme)
		deserializer  = codecs.UniversalDeserializer()
	)
	ss := &server{
		d: deserializer,
		l: log.Default(),
		c: map[string]agentConfig{
			"java": agentConfig{
				container: "agent-container",
				environment: []string{
					"ENVIRONMENT_VARIABLE1=value1",
					"ENVIRONMENT_VARIABLE2=value2",
					"ENVIRONMENT_VARIABLE3=value3",
				},
			},
		},
	}
	s := &http.Server{
		Addr:    ":8443",
		Handler: ss,
	}
	log.Println("listening on :8443")
	log.Fatal(s.ListenAndServeTLS("/webhook.pem", "/webhook.key"))
}

type server struct {
	d runtime.Decoder
	l *log.Logger
	c map[string]agentConfig
}

// TODO: Add parsing
type agentConfig struct {
	container string
	// TODO: define in config as map[string]string, but convert to []string
	// on parse.
	environment []string
}

const apmAnnotation = "elastic-apm-agent"

func (s *server) mutate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		s.l.Printf("failed to unmarshal to pod: %v\n", err)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	s.l.Printf("AdmissionReview for kind=%v namespace=%v name=%v (%v) UID=%v patchOperation=%v userInfo=%v\n",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	// create patch
	patch, err := s.createPatch(pod)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("err: %v", err),
			},
		}
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("err: %v", err),
			},
		}
	}
	// return AdmissionResponse
	patchTypeJSON := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchTypeJSON,
	}
}

type patchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (s *server) createPatch(pod corev1.Pod) ([]patchOp, error) {
	annotations := pod.ObjectMeta.GetAnnotations()
	if annotations == nil {
		return nil, errors.New("no annotations present")
	}
	// TODO: Do we want to support multiple comma-separated agents?
	agent, ok := annotations[apmAnnotation]
	if !ok {
		return nil, errors.New("missing annotation `elastic-apm-agent`")
	}
	// TODO: validate the config has a container field
	config, ok := s.c[agent]
	if !ok {
		return nil, fmt.Errorf("no config for agent `%s`", agent)
	}

	// adding each environment variable to agent container + appending
	// agent container to containers list
	patch := make([]patchOp, 0, len(config.environment)+1)
	agentIdx := len(pod.Spec.Containers) + 1

	for i, envPair := range config.environment {
		patch = append(patch, patchOp{
			Op: "add",
			// TODO: Is this right?
			Path:  fmt.Sprintf("/spec/containers/%d/env/%d", agentIdx, i),
			Value: envPair,
		})
	}
	// TODO: Is it necessary to have EACH container in the patch operation?
	// or just the one I'm adding?
	// TODO: Seems like network space within a pod is flat, ie. the app can
	// communicate on the default localhost:XXXX address to address the
	// agent in a separate container.
	patch[agentIdx] = patchOp{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/image", agentIdx),
		Value: config.container,
	}

	return patch, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// limit reader?
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading body: %v", err), http.StatusInternalServerError)
		return
	}
	// check body length?

	// verify the content type is accurate?
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *admissionv1.AdmissionResponse
	ar := admissionv1.AdmissionReview{}
	if _, _, err := s.d.Decode(body, nil, &ar); err != nil {
		errStr := fmt.Sprintf("err decoding body: %v", err)
		s.l.Println(errStr)
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	admissionResponse = s.mutate(&ar)

	admissionReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	if err := json.NewEncoder(w).Encode(admissionReview); err != nil {
		s.l.Printf("marshal admissionReview err: %v", err)
	}
}
