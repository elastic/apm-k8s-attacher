TODO:
- TLS (is this needed?)
- manifest files
  - client app w/ annotation
  - service for webhook
  - webhook config
  - what else
- KinD

process:

- start local kubernetes cluster with KinD
- create prototype mutating webhook server
- create deployment/service spec for webhook server
- create and apply `MutatingWebhookConfiguration`
  - connect via service ip
- create dummy service with annotation; dump out environment in appended dummy
  "agent" container to verify environment written and agent container started

Links:
- apm-server issue: https://github.com/elastic/apm-server/issues/7386
- apm issue: https://github.com/elastic/apm/issues/385
- [Using Admission Controllers | Kubernetes](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)
- [MutatingWebhook config options](https://pkg.go.dev/k8s.io/api/admissionregistration/v1beta1#MutatingWebhook)

apm-server has webhook endpoint, receives pod.yml, adds environment variables

- Idempotent? Ways to limit repeat calls?
- Pods have access to agent binaries and can start them?
  - Istio injects an Envoy sidecar container to target pods to implement
    traffic management and policy enforcement.
- check if TLS is required webhook running in-cluster

simple tutorial:
https://medium.com/ovni/writing-a-very-basic-kubernetes-mutating-admission-webhook-398dbbcb63ec
https://github.com/alex-leonhardt/k8s-mutate-webhook

pods opt in with a label
```
namespaceSelector:
  matchLabels:
    mutateme: enabled
```

other, possible better tutorial:
https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74
https://github.com/morvencao/kube-sidecar-injector

1. define environment variables+values for given agent when starting webhook server
2. check for annotation, eg. `elastic-apm-agent=java`
3. apply config matching annotation name
```
for _, pod := range pods {
  v, ok := pod.annotations['elastic-apm-agent']
  if !ok { return nil }
  cfg, ok := config[v] {
  if !ok { return nil }
  for _, envVar := cfg['environment'] {
    // inject env var into pod environment
  }
  // add agent container to pod, cf. istio?
}
```

yml config
```yml
agents:
  java:
    container: docker.com/elastic/agent-java:1.2.3
    environment:
      SOME_VAR1: value1
      SOME_VAR2: value2
      SOME_VAR2: value3
  node: # no environment, run with defaults
    container: docker.com/elastic/agent-node:1.2.3
```
