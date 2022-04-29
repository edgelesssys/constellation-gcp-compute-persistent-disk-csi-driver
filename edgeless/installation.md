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
kubectl apply -k ./deploy/kubernetes/overlays/edgeless/v1.0.0
```

Wait for the driver setup to finish:
```shell
kubectl wait -n kube-system csi-gce-pd-controller --for condition=available 
```

Proceed to [use](use.md) to learn how to create a storage class for provisioning encrypted storage to your workloads.


## Enabling integrity protection

By default the CSI driver will transparently encrypt all disks staged on the node.
Optionally, you can configure the driver to also apply integrity protection.

Please note that enabling integrity protection requires wiping the disk before use.
For small disks (10GB-20GB) this may only take a minute or two, while larger disks can take up to an hour or more, potentially blocking your Pods from starting for that time.
If you intend to provision large amounts of storage and Pod creation speed is important, we recommend to not use this option.

To enable integrity protection support for the CSI driver, set `--integrity` to `true` in `deploy/kubernetes/overlays/edgeless/v1.0.0/node-args.yaml` and apply the changes:
```shell
sed -i s/--integrity=false/--integrity=true/g ./deploy/kubernetes/overlays/edgeless/v1.0.0/node-args.yaml
kubectl apply -k ./deploy/kubernetes/overlays/edgeless/v1.0.0
```


## Clean up

Remove the driver from your Constellation by deleting the namespace:
```shell
kubectl delete -k ./deploy/kubernetes/overlays/edgeless/v1.0.0
```
