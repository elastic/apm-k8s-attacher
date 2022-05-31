#!/bin/bash
set -exuo errexit

export APP="example-app"
export NAMESPACE="${2:-default}"

revision=$(date +%s)

kubectl apply -f - <<EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${APP}
  template:
    metadata:
      name: ${APP}
      labels:
        app: ${APP}
      annotations:
        deployment.kubernetes.io/revision: "$revision"
    spec:
      containers:
        - name: example-app
          image: hashicorp/http-echo:alpine
          imagePullPolicy: Always
          args:
          - "-text='hello world'"
          ports:
          - containerPort: 5678
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP}-annotation
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}-annotation
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${APP}-annotation
  template:
    metadata:
      name: ${APP}-annotation
      labels:
        app: ${APP}-annotation
      annotations:
        deployment.kubernetes.io/revision: "$revision"
        co.elastic.traces/agent: java
    spec:
      containers:
        - name: example-app
          image: hashicorp/http-echo:alpine
          imagePullPolicy: Always
          args:
          - "-text='hello world'"
          ports:
          - containerPort: 5678
          env:
          - name: ELASTIC_APM_LOG_LEVEL
            value: "error"
          - name: ELASTIC_APM_SERVICE_NAME
            value: "original-name"
EOF
