---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-webhook-add-pod-annotation.html
---

# Add a pod template annotation to each pod you want to auto-instrument [apm-webhook-add-pod-annotation]

To auto-instrument a deployment, update its `spec.template.metadata.annotations` to include the `co.elastic.apm/attach` key. The webhook matches the value of this key to the `webhookConfig.agents` value defined in your Helm values file.

For example, if your Webhook values file includes the following:

```yaml
...
webhookConfig:
  agents:
    java:
...
```

Then your `co.elastic.apm/attach` value should be `java`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  # ...
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        co.elastic.apm/attach: java <1>
      labels:
        # ...
    spec:
      #...
```

1. The APM attacher configuration `webhookConfig.agents.java` matches `co.elastic.apm/attach: java`. If you define further configurations, for example the `java-dev` configuration below, and you wanted to use that definition for this deployment, this entry would be `java-dev` instead of `java`


The `spec.template.metadata.annotations` value allows you to set custom environment variables and images per deployment. For example, your Helm values file might configure a number of deployments: `java-dev` might have a different APM environment from `java-prod`, and `backend2` use a different APM agent than other deployments.

```yaml
agents:
  java-dev:
    image: docker.elastic.co/observability/apm-agent-java:latest
    artifact: "/usr/agent/elastic-apm-agent.jar"
    environment:
      ELASTIC_APM_SERVER_URL: "http://192.168.1.10:8200"
      ELASTIC_APM_ENVIRONMENT: "dev"
      ELASTIC_APM_LOG_LEVEL: "debug"
      ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED: "true"
      JAVA_TOOL_OPTIONS: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"
  java-prod:
    image: docker.elastic.co/observability/apm-agent-java:1.44.0 <1>
    artifact: "/usr/agent/elastic-apm-agent.jar"
    environment:
      ELASTIC_APM_SERVER_URL: "http://192.168.1.11:8200"
      ELASTIC_APM_ENVIRONMENT: "prod"
      ELASTIC_APM_LOG_LEVEL: "info"
      ELASTIC_APM_PROFILING_INFERRED_SPANS_ENABLED: "true"
      JAVA_TOOL_OPTIONS: "-javaagent:/elastic/apm/agent/elastic-apm-agent.jar"
  backend2:
    image: docker.elastic.co/observability/apm-agent-nodejs:latest
    artifact: "/opt/nodejs/node_modules/elastic-apm-node"
    environment:
      NODE_OPTIONS: "-r /elastic/apm/agent/elastic-apm-node/start"
      ELASTIC_APM_SERVER_URL: "http://192.168.1.11:8200"
      ELASTIC_APM_SERVICE_NAME: "petclinic"
      ELASTIC_APM_LOG_LEVEL: "info"
```

1. The example here shows a `java-prod` configuration which specifies a specific version of the agent instead of the `latest`


::::{important}
The only `webhookConfig.agents` values defined in [`values.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/charts/apm-attacher/values.yaml) are `java` and `nodejs`. When using other values, you must explicitly specify `image`, `artifact`, and `*OPTIONS` values.
::::


::::{important}
The environment variables defined in the webhook and here take precedence - overwrite - the values defined in the Kubernetes deployments. For example if your image uses JAVA_TOOL_OPTIONS, the value your image sets will be ignored in favour of the value set here or in the [`values.yaml`](https://github.com/elastic/apm-k8s-attacher/blob/main/charts/apm-attacher/values.yaml).
::::


