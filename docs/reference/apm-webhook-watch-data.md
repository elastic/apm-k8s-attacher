---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-webhook-watch-data.html
---

# Watch data flow into the Elastic Stack [apm-webhook-watch-data]

You may not see data flow into the {{stack}} right away; that’s normal. The addition of a pod annotation does not trigger an automatic restart. Therefore, existing pods will will not be affected by the APM Attacher. Only new pods—​as they are created via the natural lifecycle of a Kubernetes deployment—​will be instrumented. Restarting pods you’d like instrumented manually will speed up this process, but that workflow is too specific to individual deployments to make any recommendations.

