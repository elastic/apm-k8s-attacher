---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-webhook-install-helm.html
applies_to:
  stack:
  serverless:
    observability:
---

# Install the webhook with Helm [apm-webhook-install-helm]

Install the webhook with Helm. Pass in your `custom.yaml` configuration file created in the previous step with the `--values` flag.

```bash
helm install [name] \ <1>
  elastic/apm-attacher \
  --namespace=elastic-apm \ <2>
  --create-namespace \
  --values custom.yaml
```

1. The name for the installed helm chart in Kubernetes.
2. The APM Attacher needs to be installed in a dedicated namespace. Any pods created in the same namespace as the attacher will be ignored.


::::{note}
`helm upgrade ...` can be used to upgrade an existing installation, eg if you have a new version of the `custom.yaml` configuration file.
::::


