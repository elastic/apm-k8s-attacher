#!/bin/bash
set -exuo errexit

export APP="petclinic"
export NAMESPACE="${2:-default}"

revision=$(date +%s)

kubectl apply -f - <<EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP}-without-attach
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}-without-attach
    service: ${APP}-without-attach
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${APP}-without-attach
  template:
    metadata:
      labels:
        app: ${APP}-without-attach
        service: ${APP}-without-attach
      annotations:
        deployment.kubernetes.io/revision: "$revision"
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: ${APP}
        image: eyalkoren/pet-clinic:without-agent
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}
    service: ${APP}
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${APP}
  template:
    metadata:
      labels:
        app: {APP}
        service: ${APP}
      annotations:
        co.elastic.apm/attach: java
        deployment.kubernetes.io/revision: "$revision"
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: ${APP}
        image: eyalkoren/pet-clinic:without-agent
---
apiVersion: v1
kind: Service
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}
spec:
  type: ClusterIP
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  selector:
    service: ${APP}
EOF
