package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
		// Handler: admitFuncHandler(applyTraceSettings),
	}
	log.Println("listening on :8443")
	if f, err := os.Stat(*certPath); err != nil {
		log.Println(err)
	} else {
		fmt.Printf("%+v\n", f)
	}
	if f, err := os.Stat(*keyPath); err != nil {
		log.Println(err)
	} else {
		fmt.Printf("%+v\n", f)
	}
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
	// var patch []patchOperation
	// if !isKubeNamespace(ar.Request.Namespace) {
	// 	var err error
	// 	patch, err = s.createPatch(pod)
	// 	if err != nil {
	// 		s.l.Println("error", err)
	// 		return &admissionv1.AdmissionResponse{
	// 			Result: &metav1.Status{
	// 				Message: fmt.Sprintf("err %v", err),
	// 			},
	// 		}
	// 	}
	// }
	// if !isKubeNamespace(ar.Request.Namespace) {
	// 	patch, err := applyTraceSettings(ar.Request)
	// 	if err != nil {
	// 		return &admissionv1.AdmissionResponse{
	// 			Result: &metav1.Status{
	// 				Message: fmt.Sprintf("err: %v", err),
	// 			},
	// 		}
	// 	}
	// }

	// patchBytes, err := json.Marshal(patch)
	// if err != nil {
	// 	s.l.Println("error", err)
	// 	return &admissionv1.AdmissionResponse{
	// 		Result: &metav1.Status{
	// 			Message: err.Error(),
	// 		},
	// 	}
	// }
	// return AdmissionResponse
	// patchTypeJSON := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Status: "Success",
		},
		Allowed: true,
		// Patch:     patchBytes,
		// PatchType: &patchTypeJSON,
	}
}

type patchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (s *server) createPatch(pod corev1.Pod) ([]patchOperation, error) {
	annotations := pod.ObjectMeta.GetAnnotations()
	if annotations == nil {
		return nil, errors.New("no annotations present")
	}
	// TODO: Do we want to support multiple comma-separated agents?
	_, ok := annotations[apmAnnotation]
	if !ok {
		return nil, errors.New("missing annotation `elastic-apm-agent`")
	}
	// Create patch operations
	var patches []patchOperation

	spec := pod.Spec

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

	// // TODO: validate the config has a container field
	// config, ok := s.c[agent]
	// if !ok {
	// 	return nil, fmt.Errorf("no config for agent `%s`", agent)
	// }

	// // adding each environment variable to agent container + appending
	// // agent container to containers list
	// patch := make([]patchOp, 0, len(config.environment)+1)
	// agentIdx := len(pod.Spec.Containers) + 1

	// for i, envPair := range config.environment {
	// 	patch = append(patch, patchOp{
	// 		Op: "add",
	// 		// TODO: Is this right?
	// 		Path:  fmt.Sprintf("/spec/containers/%d/env/%d", agentIdx, i),
	// 		Value: envPair,
	// 	})
	// }
	// // TODO: Is it necessary to have EACH container in the patch operation?
	// // or just the one I'm adding?
	// // TODO: Seems like network space within a pod is flat, ie. the app can
	// // communicate on the default localhost:XXXX address to address the
	// // agent in a separate container.
	// patch[agentIdx] = patchOp{
	// 	Op:    "add",
	// 	Path:  fmt.Sprintf("/spec/containers/%d/image", agentIdx),
	// 	Value: config.container,
	// }

	// return patch, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.l.Println("request received")
	// read the body / request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	// mutate the request
	mutated, err := Mutate(body, true)
	if err != nil {
		sendError(err, w)
		return
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
	return

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

	bts, err := json.Marshal(admissionReview)
	if err != nil {
		s.l.Printf("marshal admissionReview err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(bts))
	if _, err := w.Write(bts); err != nil {
		s.l.Printf("marshal admissionReview err: %v", err)
	}
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func Mutate(body []byte, verbose bool) ([]byte, error) {
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
