# PD CSI driver (Edgeless edition) installation guide

## Prerequisites

Create a docker-registry secret to configure pull access for the driver:

```shell
kubectl create secret docker-registry regcred \
    --docker-server=DOCKER_REGISTRY_SERVER \
    --docker-username=DOCKER_USER \
    --docker-password=DOCKER_PASSWORD \
    --docker-email=DOCKER_EMAIL
    --namespace=kube-system
```

## Installation

Install the driver:

```shell
kubectl apply -k ./deploy/kubernetes/overlays/edgeless/v1.3.0
```

Wait for the driver setup to finish:

```shell
kubectl wait -n kube-system deployments csi-gce-pd-controller --for condition=available
```

Proceed to [use](use.md) to learn how to create a storage class for provisioning encrypted storage to your workloads.

## Clean up

Remove the driver from your Constellation by deleting the namespace:

```shell
kubectl delete -k ./deploy/kubernetes/overlays/edgeless/v1.3.0
```
