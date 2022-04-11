#!/bin/bash
set -exuo errexit

export APP="${1}"
export NAMESPACE="${2:-default}"

revision=$(date +%s)

kubectl apply -f - <<EOF
---
apiVersion: v1
kind: Service
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}
  annotations:
    deployment.kubernetes.io/revision: "$revision"
spec:
  publishNotReadyAddresses: true
  ports:
    - port: 443
      targetPort: 8443
  selector:
    app: ${APP}

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
    spec:
      containers:
        - name: ${APP}
          image: stuartnelson3/${APP}:latest
          imagePullPolicy: IfNotPresent
EOF
