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
	s.l.Println("request received")
	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	// mutate the request
	mutated, err := mutate(body, true)
	if err != nil {
		sendError(err, w)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func mutate(body []byte, verbose bool) ([]byte, error) {
	// unmarshal request into AdmissionReview struct
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

		// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		// set response options
		resp.Allowed = true
		resp.UID = ar.UID
		pT := admissionv1.PatchTypeJSONPatch
		resp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

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

		// parse the []map into JSON
		resp.Patch, err = json.Marshal(patches)

		// Success, of course ;)
		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp
		// back into JSON so we can return the finished AdmissionReview w/ Response directly
		// w/o needing to convert things in the http handler
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err // untested section
		}
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody)) // untested section
	}

	return responseBody, nil
}
