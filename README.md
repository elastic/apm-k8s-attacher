# installing kubectl and KinD

kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/

```
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
```

kind: https://kind.sigs.k8s.io/docs/user/quick-start/

```
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64
chmod +x ./kind
mv ./kind /some-dir-in-your-PATH/kind
```

```
kind create cluster --config kind.yaml
```

a config is created at `~/.kube/config`, which is already set to communicate
with the cluster. if using two clusters, cf.:
https://kind.sigs.k8s.io/docs/user/quick-start/#interacting-with-your-cluster

# removing KinD

get the available clusters:

```
kind get clusters
```

delete desired clusters

```
kind delete cluster <cluster-name>
```

# debugging

docker exec into the running KinD node
From there, the pod network is exposed on the host, ie.

```
docker exec -it <kind container id> bash
kubectl get pods -o wide
# note the ip addr
curl 10.244.0.16:5678
```

# developing

## helm chart

make changes to the helm chart, and then you can install/upgrade it in the
cluster:

```
helm install webhook apm-agent-auto-attach/
helm upgrade webhook apm-agent-auto-attach/
```

## webhook

Do your normal go development in the top-level *.go files of this repo.

### creating the webhook container

Note: The container used is alpine, because it's tiny but still allows for some
degree of debugging. You're pretty much sunk if you're using scratch.

1. create container: `make .webhook`
2. make the webhook is available on dockerhub. mine is already there, you'll
   have to change the container name if you want to use your own. this will
   require updating the helmchart.

## deploying the example container

to deploy a simple echo server:

```
./example_deploy.sh
```

it already has the correct annotation. you can check that it's been configured
correctly by the webhook using `kubectl`.

# old notes that probably are no longer relevant

## tls

Note: this is handled now within the helmchart

generating the tls certs for local testing are documented here; the user will
most likely be bringing their own. loading them into the cluster is also
documented; clients and testing will have to do this regardless.

1. things are currently hardcoded to use `webhook` and `default` as the app
   name and namespace. in the future, maybe you can decide on your own names,
   but for now hardcoding is easier.
2. run the generation script: `make webhook.pem`

TODO: tls cert management. they probably want to use kubectl secret store and
not have the certs in the container.

https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74#e859

## AdmissionWebhookRegistration object

```
./submit-webhook-config.sh webhook
```

## notes

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
