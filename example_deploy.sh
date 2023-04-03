#!/bin/bash
set -exuo errexit

export APP="example-app"

revision=$(date +%s)

kubectl apply -f - <<EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: petclinic-without-attach
  labels:
    app: petclinic-without-attach
    service: petclinic-without-attach
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: petclinic-without-attach
  template:
    metadata:
      labels:
        app: petclinic-without-attach
        service: petclinic-without-attach
      annotations:
        deployment.kubernetes.io/revision: "$revision"
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: petclinic
        image: eyalkoren/pet-clinic:without-agent
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: petclinic
  labels:
    app: petclinic
    service: petclinic
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: petclinic
  template:
    metadata:
      labels:
        app: petclinic
        service: petclinic
      annotations:
        co.elastic.apm/attach: java
        deployment.kubernetes.io/revision: "$revision"
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: petclinic
        image: eyalkoren/pet-clinic:without-agent
---
apiVersion: v1
kind: Service
metadata:
  name: petclinic
  namespace: default
  labels:
    app: petclinic
spec:
  type: ClusterIP
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  selector:
    service: petclinic
EOF
