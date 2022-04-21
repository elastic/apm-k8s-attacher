#!/bin/bash

set -exuo errexit

namespace=elastic-apm

go build .
make .webhook
docker push stuartnelson3/webhook
webhook_name=$(kubectl get -o name pods --namespace=$namespace)
kubectl delete $webhook_name --namespace=$namespace
sleep 1
webhook_name=$(kubectl get -o name pods --namespace=$namespace)
kubectl wait --for=condition=Ready=true $webhook_name --namespace=$namespace
kubectl logs -f $webhook_name --namespace=$namespace
