package main

import (
	"encoding/json"
	"flag"
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
		certPath      = flag.String("certFile", "/opt/webhook/certs/cert.pem", "path to cert.pem")
		keyPath       = flag.String("keyFile", "/opt/webhook/certs/key.pem", "path to key.pem")
	)
	flag.Parse()
	ss := &server{
		d: deserializer,
		l: log.Default(),
		c: map[string]agentConfig{
			"java": agentConfig{
				container: "stuartnelson3/agent-container",
				environment: []string{
					"ENVIRONMENT_VARIABLE1=value1",
					"ENVIRONMENT_VARIABLE2=value2",
					"ENVIRONMENT_VARIABLE3=value3",
				},
			},
		},
	}
	_ = ss
	s := &http.Server{
		Addr:    ":8443",
		Handler: ss,
	}
	log.Println("listening on :8443")
	log.Fatal(s.ListenAndServeTLS(*certPath, *keyPath))
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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	mutated, err := mutate(body)
	if err != nil {
		sendError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func mutate(body []byte) ([]byte, error) {
	admReview := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := admissionv1.AdmissionResponse{}

	if ar != nil {
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		resp.Allowed = true
		resp.UID = ar.UID
		pT := admissionv1.PatchTypeJSONPatch
		resp.PatchType = &pT

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

		resp.Patch, err = json.Marshal(patches)
		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err
		}
	}

	return responseBody, nil
}
