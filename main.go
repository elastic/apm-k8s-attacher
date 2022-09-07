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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var (
		certPath   = flag.String("certFile", "/opt/webhook/certs/cert.pem", "path to cert.pem")
		keyPath    = flag.String("keyFile", "/opt/webhook/certs/key.pem", "path to key.pem")
		configPath = flag.String("config", "/opt/webhook/config/webhook.yaml", "path to config.yaml")
	)
	flag.Parse()

	cfg, err := parseConfig(*configPath)
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	ss := &server{
		l: log.Default(),
		c: cfg.Agents,
	}
	s := &http.Server{
		Addr:    ":8443",
		Handler: ss,
	}
	log.Println("listening on :8443")
	log.Fatal(s.ListenAndServeTLS(*certPath, *keyPath))
}

func parseConfig(configPath string) (*config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := new(config)
	return c, yaml.NewDecoder(f).Decode(c)
}

type server struct {
	l *log.Logger
	c map[string]agentConfig
}

type config struct {
	Agents map[string]agentConfig `yaml:"agents"`
}

type agentConfig struct {
	Image        string            `yaml:"image"`
	Environment  map[string]string `yaml:"environment"`
	ArtifactPath string            `yaml:"artifact"`
}

const apmAnnotation = "co.elastic.traces/agent"

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendError(err, w)
		return
	}

	admReview := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		s.sendError(err, w)
		return
	}

	if err := s.mutate(&admReview); err != nil {
		s.sendError(err, w)
		return
	}

	resp, err := json.Marshal(admReview)
	if err != nil {
		s.sendError(err, w)
		return
	}
	if _, err := w.Write(resp); err != nil {
		s.sendError(err, w)
		return
	}
}

func (s *server) sendError(err error, w http.ResponseWriter) {
	s.l.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func (s *server) mutate(admReview *admissionv1.AdmissionReview) error {
	var pod *corev1.Pod

	ar := admReview.Request
	resp := admissionv1.AdmissionResponse{}

	if ar == nil {
		// TODO: Is this right?
		return nil
	}

	if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
		return fmt.Errorf("unable unmarshal pod json object %v", err)
	}

	resp.Allowed = true
	resp.UID = ar.UID

	config, err := getConfig(s.c, pod.ObjectMeta.GetAnnotations())
	if err != nil {
		resp.Result = &metav1.Status{Message: err.Error()}
		admReview.Response = &resp
		return nil
	}

	pT := admissionv1.PatchTypeJSONPatch
	resp.PatchType = &pT

	patch, err := createPatch(config, pod.Spec)
	if err != nil {
		return err
	}

	marshaled, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	resp.Patch = marshaled

	resp.Result = &metav1.Status{
		Status: "Success",
	}

	admReview.Response = &resp
	return nil
}

func getConfig(configs map[string]agentConfig, annotations map[string]string) (agentConfig, error) {
	ac := agentConfig{}
	if annotations == nil {
		return ac, errors.New("no annotations present")
	}
	// TODO: Do we want to support multiple comma-separated agents?
	agent, ok := annotations[apmAnnotation]
	if !ok {
		return ac, errors.New("missing annotation `co.elastic.traces/agent`")
	}
	// TODO: validate the config has a container field
	config, ok := configs[agent]
	if !ok {
		return ac, fmt.Errorf("no config for agent `%s`", agent)
	}
	return config, nil
}
