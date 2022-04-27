# apm mutating webhook

This is the repo for apm mutating webhook for kubernetes. It contains both the
webhook, and a helmchart for the webhook.

# todo

- inject kubernetes environment variables via downward API (https://www.elastic.co/guide/en/apm/guide/current/api-metadata.html#api-kubernetes-data)
- auto-discovery of apm-server
- publish "official" container image to elastic docker hub
- default + customizable webhook configuration
- configure a regex matcher on container image name that applies the same agentConfig
- well-defined paths for agent artifacts in containers
- current version of agent containers as defaults
- @felix connects with OPA team to discuss their strategy and our strategy
- add "why helm" doc
- move all these to issues on github
- update repo perms
- initial demo presentation https://drive.google.com/drive/folders/18TMg1AQ0xcIddPGmR3Ty76ODDwukmTai
- security audit
- k8s audit
- set up deployment infra (build + release container)

# demo

```
# create cluster
kind create cluster --config kind.yaml
# install helm chart
helm upgrade -i webhook apm-agent-auto-attach/ --namespace=elastic-apm --create-namespace
# add deploy with annotation
./example_deploy.sh
# query for pod name
pod=$(kubectl get -o name pods | grep annotation)
# verify it has been mutated (environment, volume)
kubectl describe $pod
```

asciinema demo: https://asciinema.org/a/aNUJAK1KUfuZwgFm4eCpOuCro

setting a custom webhook config:

Given a file `custom.yaml`:
```yaml
webhookConfig:
  agents:
    java:
      image: docker.elastic.co/observability/apm-agent-java:1.23.0
      environment:
        ELASTIC_APM_SERVER_URLS: "http://34.78.173.219:8200"
        ELASTIC_APM_SERVICE_NAME: "custom"
        ELASTIC_APM_ENVIRONMENT: "dev"
        ELASTIC_APM_LOG_LEVEL: "debug"
        ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED: "true"
        JAVA_TOOL_OPTIONS: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"
```

The user can inject their own custom config for the mutating webhook:
```
helm upgrade -i webhook apm-agent-auto-attach/ --namespace=elastic-apm --create-namespace -f custom.yaml
```

The annotation looked for on a pod is `elastic-apm-agent`. The value indicates
which image+environment variables to inject into the pod. eg., `java` would
inject the java image + environment variables, `node` would inject the node
image + environment variables. the actual value is unimportant, it's just the
config that it contains that matters.

the user also needs to define `apm.token`. This can be written in either the
`custom.yaml` file, or applied via `--set apm.token=$MY_TOKEN` when running
`helm`.

# configuring

The user can (and should) pass in a custom yml config on creation. How do we
want to handle this? Do we provide an example they should use and update
themselves? Server url, service name, etc are not things we can provide a
default for, right?

```yml
agents:
  java:
    image: docker.com/elastic/agent-java:1.2.3
    artifact: "/usr/agent/elastic-apm-agent.jar"
    environment:
      ELASTIC_APM_SERVER_URLS: "http://34.78.173.219:8200"
      ELASTIC_APM_SERVICE_NAME: "petclinic"
      ELASTIC_APM_ENVIRONMENT: "test"
      ELASTIC_APM_LOG_LEVEL: "debug"
      ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED: "true"
      JAVA_TOOL_OPTIONS: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"
  node: # no environment, run with defaults
    image: docker.com/elastic/agent-node:1.2.3
```

Using the annotation value allows users to set custom environment variables and
images per deploy. For example, `backend1` might have a different service name
from `backend2`, and `backend1-dev` might have a different apm environment from
`backend1-prod`.

Note: Right now, we can only specify a single secret token to be injected for
interacting with the apm-server, which means a single webhook deploy can only
configure pods to one apm-server. A user can install multiple versions of the
helm chart, however, with different apm-server/secret-token combinations, and
have different values for the agent configs. Coming up with a way to configure
this so that a token can be related to a specific apm-server might take some
additional thought.

Open questions:
- How do we configure the command for moving the agent artifact into the shared
  volume? Right now it's using `artifact` (see example config above) to know
  the location, and then copying it to the non-configurable
  `/elastic/apm/agent/$ARTIFACT` within the shared volume in the pod.
- Currently the artifact location (see `JAVA_TOOL_OPTIONS` above) is hardcoded
  within `patch.go`. The `JAVA_TOOL_OPTIONS` environment variable depends on a
  constant within the code; how can we prevent a user from configuring this
  incorrectly? Or is this something we "supply" in a default config and hope
  they don't mess with it?

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
helm upgrade -i webhook apm-agent-auto-attach/ --namespace=elastic-apm --create-namespace
helm uninstall webhook --namespace=elastic-apm
```

## webhook

Do your normal go development in the top-level *.go files of this repo.

### creating the webhook container

Note: The container used is alpine, because it's tiny but still allows for some
degree of debugging. You're pretty much sunk if you're using scratch.

1. create container: `make .webhook`
2. make the webhook is available on dockerhub. mine is already there, you'll
   have to change the container name if you want to use your own. this will
   require updating the helmchart. you'll have to update the helm variable
   `image.repository` to your repo, `--set image.repository=$MY_REPO`.

## deploying the example container

to deploy a simple echo server:

```
./example_deploy.sh
```

it already has the correct annotation. you can check that it's been configured
correctly by the webhook using `kubectl`.

# notes

Links:
- apm-server issue: https://github.com/elastic/apm-server/issues/7386
- apm issue: https://github.com/elastic/apm/issues/385
- [Using Admission Controllers | Kubernetes](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)
- [MutatingWebhook config options](https://pkg.go.dev/k8s.io/api/admissionregistration/v1beta1#MutatingWebhook)

source code inspiration:
https://github.com/ExpediaGroup/kubernetes-sidecar-injector/tree/master

simple tutorial:
https://medium.com/ovni/writing-a-very-basic-kubernetes-mutating-admission-webhook-398dbbcb63ec
https://github.com/alex-leonhardt/k8s-mutate-webhook

other tutorial:
https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74
https://github.com/morvencao/kube-sidecar-injector

