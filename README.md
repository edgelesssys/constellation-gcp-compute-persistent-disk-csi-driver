# Google Compute Engine Persistent Disk CSI Driver

- [Upstream source](https://github.com/kubernetes-sigs/azuredisk-csi-driver)
- [Constellation repo](https://github.com/edgelesssys/constellation)

## About

This driver allows a Constellation cluster to use [GCP Persistent Disks](https://cloud.google.com/persistent-disk).

## Plugin Features

### CreateVolume Parameters

| Parameter        | Values                    | Default       | Description                                                                                        |
|------------------|---------------------------|---------------|----------------------------------------------------------------------------------------------------|
| type             | Any PD type (see [GCP documentation](https://cloud.google.com/compute/docs/disks#disk-types)), eg `pd-ssd` `pd-balanced` | `pd-standard` | Type allows you to choose between standard Persistent Disks  or Solid State Drive Persistent Disks |
| replication-type | `none` OR `regional-pd`   | `none`        | Replication type allows you to choose between Zonal Persistent Disks or Regional Persistent Disks  |
| disk-encryption-kms-key | Fully qualified resource identifier for the key to use to encrypt new disks. | Empty string. | Encrypt disk using Customer Managed Encryption Key (CMEK). See [GKE Docs](https://cloud.google.com/kubernetes-engine/docs/how-to/using-cmek#create_a_cmek_protected_attached_disk) for details. |
| labels           | `key1=value1,key2=value2` |               | Labels allow you to assign custom [GCE Disk labels](https://cloud.google.com/compute/docs/labeling-resources). |

### Topology

This driver supports only one topology key:
`topology.gke.io/zone`
that represents availability by zone (e.g. `us-central1-c`, etc.).

## Driver Deployment

Install the driver using kubectl:

```shell
kubectl apply -k ./deploy/kubernetes/overlays/edgeless/latest
```

Wait for the driver setup to finish:

```shell
kubectl wait -n kube-system deployments csi-gce-pd-controller --for condition=available
```

Proceed to [use](edgeless/use.md) to learn how to create a storage class for provisioning encrypted storage to your workloads.

Remove the driver using kubectl:

```shell
kubectl delete -k ./deploy/kubernetes/overlays/edgeless/latest
```

## Further Documentation

- [Local Development](docs/local-development.md)
- [User Guides](docs/kubernetes/user-guides)
- [Driver Development](docs/kubernetes/development.md)

To build the driver container image:

```shell
driver_version=v0.0.0-test
GCE_PD_CSI_STAGING_IMAGE=ghcr.io/edgelesssys/encrypted-gcp-csi-driver \
  GCE_PD_CSI_STAGING_VERSION=${driver_version} \
  make push-container
```
