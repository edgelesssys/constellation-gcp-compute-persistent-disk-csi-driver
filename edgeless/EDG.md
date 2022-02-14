# Constellation modifications & documentation

## Reqiured permissions

Permissions required by the service account:
```
compute.instances.get
compute.instances.attachDisk
compute.instances.detachDisk
roles/compute.storageAdmin
roles/iam.serviceAccountUser
```

Permissions required to create a service account with the required permissions:
```
resourcemanager.projects.getIamPolicy
resourcemanager.projects.setIamPolicy
iam.serviceAccounts.create
iam.serviceAccounts.delete
```

Permissions required to run the setup script:
```
iam.serviceAccounts.list
iam.serviceAccountKeys.create
iam.roles.create
iam.roles.get
iam.roles.update
```

## Deploying the driver

A service account is required to use the storage driver.
If no service account exists the script `./deploy/setup-project.sh` can be used to create the service account.
```shell
PROJECT=<PROJECT_ID> \
    GCE_PD_SA_NAME=<SERVICE_ACCOUNT_NAME> \
    GCE_PD_SA_DIR=</directory/for/credentials> \
    ENABLE_KMS=<true/false> \
    ./deploy/setup-project.sh
```

Deploy the driver to the Cluster.
```
GCE_PD_SA_DIR=</directory/for/credentials> \
    GCE_PD_DRIVER_VERSION=edgeless \
    ./deploy/kubernetes/deploy-driver.sh
```

Create a storage class:
```yaml
kubectl apply -f edgeless/test-sc.yaml
```

We can now create PersistentVolumeClaims using `storageClassName: csi-gce-pd`:

```shell
cat <<EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: podpvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: csi-gce-pd
  resources:
    requests:
      storage: 20Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: web-server
spec:
  containers:
   - name: web-server
     image: nginx 
     volumeMounts:
       - mountPath: /var/lib/www/html
         name: mypvc
  volumes:
   - name: mypvc
     persistentVolumeClaim:
       claimName: podpvc
       readOnly: false
EOF
```

## Cleanup

Remove the driver by running the following:
```
./deploy/kubernetes/delete-driver.sh
```

## Build your own driver

Build and push container:
```shell
GCE_PD_CSI_STAGING_IMAGE=ghcr.io/edgelesssys/gcp-csi-driver \
  GCE_PD_CSI_STAGING_VERSION=edgeless \
  make push-container
```

Create a pull secret for the storage driver (only necessary if pulling from a private repository):
```shell
kubectl create namespace gce-pd-csi-driver
kubectl create secret generic regcred \
   --namespace=gce-pd-csi-driver \
  --from-file=.dockerconfigjson=</path/to/.docker/config.json> \
  --type=kubernetes.io/dockerconfigjson
```

Replace `gke.gcr.io/gcp-compute-persistent-disk-csi-driver` in `deploy/images/stable-master/image.yaml` with your own image and tag.
You should now be able to deploy our version of the storage driver: 
```shell
GCE_PD_SA_DIR=</directory/for/credentials> \
    GCE_PD_DRIVER_VERSION=stable-master \
    ./deploy/kubernetes/deploy-driver.sh
```

## Storage compatibility

This storage driver plugin only supports Google's persistent-disk storage. This type of storage can only be mounted using SCSI.

For confidential VMs we need storage to be mounted over NVMe (available for local SSD), which is not supported by the plugin.
See [this document](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/local-ssd) for information about using local SSD in GKE.
However, local SSD is disabled for confidential GKE nodes. Google's own CSI driver also does not work with confidential nodes.
