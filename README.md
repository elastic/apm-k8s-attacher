# apm mutating admission webhook

This is the repo for apm mutating admission webhook for kubernetes. It contains
both the webhook receiver and a helmchart for managing the receiver's
lifecycle within kubernetes.

The software contained in this repo is considered a **technical preview**, and
is not yet recommended for use in a production cluster.

## webhook

The purpose of the webhook receiver is to modify pods so that they are
automatically instrumented by an elastic apm agent. Currently, the java,
nodejs, and dotnet agents are supported.

The webhook receiver is invoked on pod creation. After having received the
object definition from the kubernetes api server, it looks through the pod spec
for a specific annotation (added by the user) and, if found, mutates the spec
according to the webhook receiver's configuration. This mutated object is then
returned to the kubernetes api server, which uses it as the source of truth for
the object.

The mutation taking place is:
1. Add an init container image, which has the agent binary
2. Add a shared volume which is mounted into both the init container image and
   all container images contained in the original incoming object
3. Copy the agent binary from the init container image into the shared volume,
   making it available to the other container images
4. Update the environment variables in the container images to configure
   auto-instrumentation with the copied agent binary

## helmchart

The helmchart manages configuring all the associated manifest files for the
webhook receiver, including generating certificates for securing communication
between the kubernetes api server and the webhook receiver.

# using the webhook

The webhook is managed by the helmchart in this repo. To install it into your
cluster, clone this repo:

```bash
git clone git@github.com:elastic/apm-mutating-webhook.git
cd apm-mutating-webhook
```

The webhook is installed by using a Helm Chart, 
you can provide a custom webhook configurations using a Helm values file. For example, editing `custom.yaml`:

```yaml
apm:
  secret_token: SuP3RT0K3N
#   api_key: VnVhQ2ZHY0JDZGJrUW0tZTVhT3g6dWkybHAyYXhUTm1zeWFrdzl0dk5udw==
#   # Custom arrays overwrite those defined in the values.yaml
#   # In this case, `[default]` is overwritten for `[ns1, ns2]`
  namespaces:
    - default
    - my-name-space-01
    - my-name-space-02
webhookConfig:
  agents:
    java:
      image: docker.elastic.co/observability/apm-agent-java:1.30.1
      artifact: "/usr/agent/elastic-apm-agent.jar"
      environment:
        JAVA_TOOL_OPTIONS: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"
        ELASTIC_APM_SERVER_URL: "https://10.10.10.10:8200"
        ELASTIC_APM_ENVIRONMENT: "prod"
        ELASTIC_APM_LOG_LEVEL: "info"
```

Modify the value of `ELASTIC_APM_SERVER_URL` in `custom.yml` to point to your
apm-server. Additionally, if you have configured API Key or secret token,
set its value as well under `apm.api_key` or `apm.secret_token` respectively.
If you're using an API Key or secret token, you need to also list all the
namespaces where you are auto-instrumenting pods. The API Key and secret token
are stored in Kubernetes `Secret` in each namespace.

Note: `artifact` and `JAVA_TOOL_OPTIONS` keys should not be edited.

To use the custom config when installing the webhook, supply `--values custom.yaml`
to the `helm upgrade` above.

Now, install the helmchart using helm:

```bash
helm upgrade \
  --install webhook apm-agent-auto-attach/ \
  --namespace=elastic-apm \
  --create-namespace \
  --values custom.yaml
```

For a deployment to be auto-instrumented, update its
`spec.template.metadata.annotations` to include `co.elastic.traces/agent: java`. The
webhook matches the value of `co.elastic.traces/agent` (in this case, `java`) to the
config with the matching name under `webhookConfig.agents` defined in the
helmchart.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
  namespace: default
  labels:
    app: my-service
    service: my-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-service
  template:
    metadata:

      # APM Mutating WebHook configuration
      annotations:
        co.elastic.traces/agent: java

      labels:
        app: my-service-java
        service: my-service
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: my-service
        image: my-service:v1.0.0
        ports:
        - name: my-service
          containerPort: 8080
```

# dev dependencies

webhook:
- golang

helmchart:
- kubectl
- kind
- helm
- skaffold
- docker

## kubectl

https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/

```
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  sudo install kubectl
```

## kind

https://kind.sigs.k8s.io/docs/user/quick-start/

```
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64 && \
  sudo install kind
```

## helm

https://github.com/helm/helm/releases

## skaffold

skaffold:
```
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && \
  sudo install skaffold /usr/local/bin/
```

# dev workflow

## webhook

Do your normal go development in the top-level *.go files of this repo.

## helmchart

start the local kubernetes cluster using `kind`:

```
kind create cluster --config kind.yaml
```

a config is created at `~/.kube/config`, which is already set to communicate
with the cluster. if using two clusters, cf.:
https://kind.sigs.k8s.io/docs/user/quick-start/#interacting-with-your-cluster

`skaffold` manages installing, updating, and removing the helmchart.

start the watcher in a separate terminal with `skaffold dev`. this watches for
changes the files within the helmchart, the Dockerfile, and any file
dependencies specified by the Dockerfile. A change will trigger an update
within the kubernetes cluster.

## debugging

docker exec into the running KinD node
From there, the pod network is exposed on the host, ie.

```
docker exec -it <kind container id> bash
kubectl get pods -o wide
# note the ip addr
curl 10.244.0.16:5678
```

## deploying the example container

to deploy a simple echo server:

```
./example_deploy.sh
```

it already has the correct annotation. you can check that it's been configured
correctly by the webhook using `kubectl`.

# removing KinD

get the available clusters:

```
kind get clusters
```

delete desired clusters

```
kind delete cluster <cluster-name>
```
