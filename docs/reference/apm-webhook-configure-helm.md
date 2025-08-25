---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-webhook-configure-helm.html
---

# Configure the webhook with a Helm values file [apm-webhook-configure-helm]

The APM Attacher’s webhook can be installed from a Helm chart. You can provide a custom webhook configuration using a Helm values file. Elastic provides a [`custom.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/custom.yaml) file as a starting point.

This sample `custom.yaml` file instruments a pod with the **Elastic APM Java agent**:

```yaml
apm:
  secret_token: SuP3RT0K3N <1>
  namespaces: <2>
    - default
    - my-name-space-01
    - my-name-space-02
webhookConfig:
  agents:
    java: <3>
      environment:
        ELASTIC_APM_SERVER_URL: "https://apm.example.com:8200" <4>
        ELASTIC_APM_ENVIRONMENT: "prod"
        ELASTIC_APM_LOG_LEVEL: "info"
```

1. The `secret_token` for your deployment. Use `api_key` if using an API key instead.
2. If you’re using a secret token or API key to secure your deployment, you must list all of the namespaces where you want to auto-instrument pods. The secret token or API key will be stored as Kubernetes Secrets in each namespace.
3. Fields written here are merged with pre-existing fields in [`values.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/charts/apm-attacher/values.yaml)
4. Elastic APM agent environment variables—for example, the APM Server URL, which specifies the URL and port of your APM integration or server.


This sample `custom.yaml` file instruments a pod with the **Elastic APM Node.js agent**:

```yaml
apm:
  secret_token: SuP3RT0K3N <1>
  namespaces: <2>
    - default
    - my-name-space-01
    - my-name-space-02
webhookConfig:
  agents:
    nodejs: <3>
      environment:
        ELASTIC_APM_SERVER_URL: "https://apm.example.com:8200" <4>
        ELASTIC_APM_ENVIRONMENT: "prod"
        ELASTIC_APM_LOG_LEVEL: "info"
```

1. The `secret_token` for your deployment. Use `api_key` if using an API key instead.
2. If you’re using a secret token or API key to secure your deployment, you must list all of the namespaces where you want to auto-instrument pods. The secret token or API key will be stored as Kubernetes Secrets in each namespace.
3. Fields written here are merged with pre-existing fields in [`values.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/charts/apm-attacher/values.yaml)
4. Elastic APM agent environment variables—for example, the APM Server URL, which specifies the URL and port of your APM integration or server.


::::{tip}
The examples above assume that you want to use the latest version of the Elastic APM agent. Advanced users may want to pin a version of the agent or provide a custom build. To do this, set your own `image`, `artifact`, and `environment.*OPTIONS` fields. Copy the formatting from [`values.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/charts/apm-attacher/values.yaml).
::::


::::{note}
Expiring and rotating API keys will need to update the `custom.yaml`, upgrade the helm install with the new `custom.yaml`, and cycle running pods in a similar way to other deployment definition changes.
::::
