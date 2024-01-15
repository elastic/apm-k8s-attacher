# How to release apm-k8s-attacher

0. Make sure everything is working by testing "main". (TODO: Clarify a manual testing procedure if one is required beyond automated tests.)
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
