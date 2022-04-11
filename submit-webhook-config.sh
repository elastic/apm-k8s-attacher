#!/bin/bash
set -exuo errexit

export APP="${1}"
export NAMESPACE="${2:-default}"

CA_BUNDLE=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')

if [[ $CA_BUNDLE = "" ]]; then
  echo "failed to get ca bundle"
  exit 1
fi

kubectl create -f - <<EOF
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: webhook
webhooks:
  - name: ${APP}.${NAMESPACE}.svc.cluster.local
    clientConfig:
      caBundle: ${CA_BUNDLE}
      service:
        name: ${APP}
        namespace: ${NAMESPACE}
        path: "/"
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    namespaceSelector:
      matchExpressions:
        - {key: elastic-apm-agent, operator: Exists}
    admissionReviewVersions: ["v1", "v1beta1"]
    sideEffects: None
EOF
