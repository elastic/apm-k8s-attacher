---
mapped_pages:
  - https://www.elastic.co/guide/en/apm/attacher/current/apm-attacher.html
  - https://www.elastic.co/guide/en/apm/attacher/current/index.html
---

:::{important}
The APM Attacher for Kubernetes is in maintenance mode. Only prioritized bug fixes will be applied. The OpenTelemetry Operator fully replaces the Attacher and is actively developed by the community.

To migrate, see [Migrate from the Elastic APM Attacher for Kubernetes](opentelemetry://reference/use-cases/kubernetes/instrumenting-applications.md#migrate-from-the-elastic-apm-attacher-for-kubernetes).
:::

# APM Attacher for Kubernetes [apm-attacher]

The APM attacher for Kubernetes simplifies the instrumentation and configuration of your application pods.

The attacher includes a [webhook receiver](#apm-webhook) that modifies pods so they are automatically instrumented by an Elastic APM agent, and a [Helm chart](#apm-helm-chart) that manages its lifecycle within Kubernetes.

Learn more below, or skip ahead to [*Instrument and configure pods*](/reference/apm-get-started-webhook.md).


## Webhook [apm-webhook]

The webhook receiver modifies pods so they are automatically instrumented by an Elastic APM agent. Supported agents include:

* [Java agent](apm-agent-java://reference/index.md)
* [Node.js agent](apm-agent-nodejs://reference/index.md)
* [preview] [.NET agent](apm-agent-dotnet://reference/index.md)

The webhook receiver is invoked on pod creation. After receiving the object definition from the Kubernetes API server, it looks through the pod spec for a specific, user-supplied annotation. If found, the pod spec is mutated according to the webhook receiver’s configuration. This mutated object is then returned to the Kubernetes API server which uses it as the source of truth for the object.


## Mutation [apm-mutation]

The mutation that occurs is defined below:

1. Add an init container image that has the agent binary.
2. Add a shared volume that is mounted into both the init container image and all container images contained in the original incoming object.
3. Copy the agent binary from the init container image into the shared volume, making it available to the other container images.
4. Update the environment variables in the container images to configure auto-instrumentation with the copied agent binary

::::{tip}
To learn more about mutating webhooks, see the [Kubernetes Admission controller documentation](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/).
::::



## Helm chart [apm-helm-chart]

The Helm chart manages the configuration of all associated manifest files for the webhook receiver, including generating certificates for securing communication between the Kubernetes API server and the webhook receiver.

::::{tip}
To learn more about Helm charts, see the [Helm documentation](https://helm.sh/docs/).
::::


