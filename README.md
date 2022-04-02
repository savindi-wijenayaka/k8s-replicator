# Kubernetes Replicator

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Main Branch Build](https://github.com/nadundesilva/k8s-replicator/actions/workflows/build-branch.yaml/badge.svg)](https://github.com/nadundesilva/k8s-replicator/actions/workflows/build-branch.yaml)

Replicator for Kubernetes resources across namespaces. This controller was written keeping extensibility in mind. Therefore, it can be extended to any other resource as needed. The following resources are supported by the Kubernetes replicator.

* Secrets

## How to Use

### How to Setup Controller

#### Quickstart

Run the following command to apply the controller to your cluster. The `<VERSION>` should be replaced with the release version
to be used and kubectl CLI should be configured pointing to the cluter in which the controller needs to be started.

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/install.sh | bash -s <VERSION>
```

#### Manual Applying

* Clone this repository.
* Update the configuration (`<REPOSITORY_ROOT>/kustomize/config.yaml`) to match your needs.
* Apply the controller into your cluster by running the following command.

  ```bash
  kubectl apply -k kustomize
  ```

### How to mark a object to be replicated

Use the following label to mark the object to be replicated.

```properties
replicator.nadundesilva.github.io/object-type=source
```

All objects with the above label will replicated into all namespaces.

### Ignored namespaces

The following namespaces are ignored by default.

* The namespace in which controller resides
* Namespaces with the name starting with `kube-` prefix
* Namespaces with the label
  ```properties
  replicator.nadundesilva.github.io/namespace-type=ignored
  ```

If you want to override this behavior and specifically replicate to a namespace, add the following label

```properties
replicator.nadundesilva.github.io/namespace-type=managed
```

### Additional labels used by the controller

The folloing labels are used by the controller to track the replication of resources.

* The following label with the value `replica` is used to mark the replicated objects.
  ```properties
  replicator.nadundesilva.github.io/object-type=replica
  ```
* The following label is used to mark a replicated resource's source namespace.
  ```properties
  replicator.nadundesilva.github.io/source-namespace=<namespace>
  ```

### How to Cleanup Controller

#### Quick Remove

Run the following command to remove the controller from your cluster. Kubectl CLI should be configured pointing to the cluter in which the controller needs to be started.

**Note:** This approach would only work if you used the Quickstart option for setting up the controller.

```bash
curl -L https://raw.githubusercontent.com/nadundesilva/k8s-replicator/main/installers/uninstall.sh | bash -s
```

#### Manual Removal

* Clone this repository.
* Remove the controller into your cluster by running the following command.

  ```bash
  kubectl delete -k kustomize
  ```

## Support

:grey_question: If you need support or have a question about the K8s Replicator, reach out through [Discussions](https://github.com/nadundesilva/k8s-replicator/discussions).

:bug: If you have found a bug and would like to get it fixed, try opening a [Bug Report](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FBug&template=bug-report.md).

:bulb: If you have a new idea or want to get a new feature or improvement added to the K8s Replicator, try creating a [Feature Request](https://github.com/nadundesilva/k8s-replicator/issues/new?labels=Type%2FFeature&template=feature-request.md).
