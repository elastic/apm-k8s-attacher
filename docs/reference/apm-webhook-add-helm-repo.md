---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-webhook-add-helm-repo.html
applies_to:
  stack:
  serverless:
    observability:
products:
  - id: cloud-serverless
  - id: observability
  - id: apm
---

# Add the helm repository to Helm [apm-webhook-add-helm-repo]

Add the Elastic helm chart repository to helm:

```bash
helm repo add elastic https://helm.elastic.co
```

