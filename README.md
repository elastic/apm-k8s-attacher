# apm mutating admission webhook

This is the repo for apm mutating admission webhook for kubernetes. It contains
both the webhook receiver and a helmchart for managing the receiver's
lifecycle within kubernetes.

The software contained in this repo is considered a **technical preview**, and
is not yet recommended for use in a production cluster.

Learn more and get started in our [documentation](elastic.co/guide/en/apm/guide/current/apm-mutating-admission-webhook.html).

## webhook

The purpose of the webhook receiver is to modify pods so that they are
automatically instrumented by an elastic apm agent. Currently, the Java and
Node.js agents are supported.

## helmchart

The helmchart manages configuring all the associated manifest files for the
webhook receiver, including generating certificates for securing communication
between the kubernetes api server and the webhook receiver.

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
