# PD CSI driver (Edgeless edition) installation guide

## Requirements

The driver relies on a service account to provision disks and make request to the Google API.
The following permissions are required by the service account:
```
compute.instances.get
compute.instances.attachDisk
compute.instances.detachDisk
roles/compute.storageAdmin
roles/iam.serviceAccountUser
```

To create the service account the following permissions are required:
```
resourcemanager.projects.getIamPolicy
resourcemanager.projects.setIamPolicy
iam.serviceAccounts.create
iam.serviceAccounts.delete
```

You can use a pre-existing service account with these roles by downloading the service account key:
```
gcloud iam service-accounts keys create "/my/safe/credentials/directory/cloud-sa.json" --iam-account "${my-iam-name}" --project "${my-project-name}"
```

Otherwise you will have to create the account manually, or use the provided script to handle the creation for you.
The following will create a service account with the driver's required permissions, and save the account key to `/my/safe/credentials/directory/cloud-sa.json`:
```shell
PROJECT="${my-project-name}" \
    GCE_PD_SA_NAME="${my-iam-name}" \
    GCE_PD_SA_DIR=/my/safe/credentials/directory \
    ./deploy/setup-project.sh
```

## Installation

Start by creating the deployment namespace:
```shell
kubectl create namespace constellation-csi-gcp
```

Create a secret holding the service account key:
```shell
kubectl create secret generic cloud-sa --from-file="/my/safe/credentials/directory/cloud-sa.json" --namespace=constellation-csi-gcp
```

[Only needed when pulling from a private repository] Create a pull secret:
```shell
kubectl create secret docker-registry regcred \
    --docker-server=DOCKER_REGISTRY_SERVER \
    --docker-username=DOCKER_USER \
    --docker-password=DOCKER_PASSWORD \
    --docker-email=DOCKER_EMAIL
    --namespace=constellation-csi-gcp
```

Install the driver:
```shell
kubectl apply -k ./deploy/kubernetes/overlays/edgeless
```

Wait for the driver setup to finish:
```shell
kubectl wait -n constellation-csi-gcp deployment csi-gce-pd-controller --for condition=available 
```

Proceed to [use](use.md) to learn how to create a storage class for provisioning encrypted storage to your workloads.

## Clean up

Remove the driver from your Constellation by deleting the namespace:
```shell
kubectl delete namespace constellation-csi-gcp
```
