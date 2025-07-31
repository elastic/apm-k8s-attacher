> [!IMPORTANT]
> The APM Attacher for Kubernetes is in maintenance mode. Only prioritized bug fixes will be applied. The OpenTelemetry Operator fully replaces the Attacher and is actively developed by the community.
> 
> To migrate, see [Migrating from the Elastic APM Attacher for Kubernetes](https://www.elastic.co/docs/reference/opentelemetry/use-cases/kubernetes/instrumenting-applications#migrate-from-the-elastic-apm-attacher-for-kubernetes).

# Elastic APM Attacher for Kubernetes

The Elastic APM attacher for Kubernetes simplifies the instrumentation and configuration of your application pods.
The attacher contains a webhook receiver and a helm chart that manages the receiver's lifecycle within kubernetes.

## Release State

The Elastic APM Attacher for Kubernetes is GA for Java and Node.js.
Support for .NET is still in **Technical Preview**, and is not yet recommended for use in a production cluster.

## Documentation

See [Elastic APM Attacher](https://www.elastic.co/guide/en/apm/attacher/current/apm-attacher.html) to get started.

## Getting Help

If you find a bug, please [report an issue](https://github.com/elastic/apm-k8s-attacher/issues).
For any other assistance, please open or add to a topic on the [APM discuss forum](https://discuss.elastic.co/c/apm).

## Contributing

See [contributing](CONTRIBUTING.md) for details about reporting bugs, requesting features, or code contributions.

## License

Apache 2.0.
