# How to release apm-k8s-attacher

0. Make sure everything is working by testing "main" (test procedure below).
1. Create a PR for the release (named "release N.M.P" or whatever):
    - Update the `version:` at "./charts/apm-attacher/Chart.yaml", e.g. "1.2.3".
    - Update the `image.tag:` at "./charts/apm-attacher/values.yaml", e.g. "v1.2.3".
      Note that this file includes a "v" prefix in the version.
    Get the PR approved and merged.
2. Working in a clone of the actual repo (not a fork), lightweight tag the repo:
    ```
    git tag vN.M.P
    git push origin vN.M.P
    ```
3. Sanity check that the release worked:
    - The release CI should trigger on the pushed tag. Check https://github.com/elastic/apm-k8s-attacher/actions/workflows/release.yml
    - https://github.com/elastic/apm-k8s-attacher/releases should show the new release.
    - The Elastic Docker registry should show the new `docker.elastic.co/observability/apm-attacher:vN.M.P` version
      and the "latest" tag should pull the same digest
        ```
        docker pull docker.elastic.co/observability/apm-attacher:vN.M.P
        docker pull docker.elastic.co/observability/apm-attacher:latest  # same digest?
        ```
    - The Elastic Helm repository should show the new release, though it may take a while (an hour?) to show up:
        ```
        helm repo add elastic https://helm.elastic.co
        helm repo update elastic
        helm search repo -l elastic/apm-attacher
        ```

## Testing procedure (for Linux, Windows works too but you need to adjust how files are created or create them manually)

1. Clone (or update) the repo locally
    - `git clone https://github.com/elastic/apm-k8s-attacher.git`
2. `cd apm-k8s-attacher`
3. Create the custom values ymal file - replacing the secret token and url with valid values for a server is better, buy even with these dummy values testing still works, just the agent won't connect to a server
```
cat > custom.yaml <<EOF
apm:
  secret_token: SuP3RT0K3N 
  namespaces: 
    - default
webhookConfig:
  agents:
    java: 
      environment:
        ELASTIC_APM_SERVER_URL: "https://apm-example.com:8200" 
        ELASTIC_APM_ENVIRONMENT: "prod"
        ELASTIC_APM_LOG_LEVEL: "info"
EOF
```
4. Install the attacher - note the namespace can be any namespace but can't be the default nor a namespace where pods will be tested
    - `helm install test-main charts/apm-attacher --values custom.yaml --namespace=elastic-apm --create-namespace` 
5. Create a pod to test - the example here is a Java pod which uses a known image that holds a testing app
```
cat > test-app.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: test-app
  annotations:
    co.elastic.apm/attach: java 
  labels:
    app: test-app
spec:
  containers:
    - image: docker.elastic.co/demos/apm/k8s-webhook-test
      imagePullPolicy: Always
      name: test-app
      env: 
      - name: ELASTIC_APM_TRACE_METHODS
        value: "test.Testing#methodB"
EOF
```
6. Start the app and check the logs
    - `kubectl apply -f test-app.yaml`
    - `kubectl logs test-app`
7. Cleanup
    - `kubectl delete -f test-app.yaml`
    - `helm delete test-main -n elastic-apm`
