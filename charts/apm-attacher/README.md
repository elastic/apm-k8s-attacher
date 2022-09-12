# APM Agent Auto Attach Helm Chart

[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/elastic)](https://artifacthub.io/packages/helm/elastic/apm-attacher)

The APM mutating admission webhook for Kubernetes simplifies the instrumentation and
configuration of your application pods.

The webhook includes a webhook receiver that modifies pods so they are automatically instrumented
by an Elastic APM agent, and Helm chart that manages the webhook receiver's lifecycle within Kubernetes.

For more information about the APM Agent Auto attacher, see:

- [Documentation](https://elastic.co/guide/en/apm/guide/current/apm-mutating-admission-webhook.html)
- [GitHub repo](https://github.com/elastic/apm-mutating-webhook)

## Requirements

<!-- TODO Kubernetes versions -->
<!-- - Supported Kubernetes versions are listed in the documentation: ? -->
- Helm >= 3.2.0

## Usage

Refer to the documentation at <https://www.elastic.co/guide/en/apm/guide/current/apm-get-started-webhook.html#apm-webhook-add-helm-repo>
