# Use


## Create a new storage class

The following will create a storage class for the CSI driver, provisioning storage of type `pd-standard` when requested.
```shell
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: encrypted-storage
provisioner: gcp.csi.confidential.cloud
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: pd-standard
EOF
```

## Make use of encrypted storage

Now you can create persistent volume claims requesting storage over your newly created storage class.
The following creates a persistent volume claim using the `encrypted-storage` class, and a Pod mounting said storage into a container:
```shell
cat <<EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: podpvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: encrypted-storage
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

## Enable integrity protection

By default the CSI driver will transparently encrypt all disks staged on the node.
Optionally, you can configure the driver to also apply integrity protection.

Please note that enabling integrity protection requires wiping the disk before use.
For small disks (10GB-20GB) this may only take a minute or two, while larger disks can take up to an hour or more, potentially blocking your Pods from starting for that time.
If you intend to provision large amounts of storage and Pod creation speed is important, we recommend to not use this option.

To enable integrity protection, create a storage class with an explicit file system type request and the integrity suffix.
The following is a storage class for integrity protected `ext4` formatted disks:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: integrity-protected
provisioner: gcp.csi.confidential.cloud
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: pd-standard
  csi.storage.k8s.io/fstype: ext4-integrity
```

## [Optional] Mark the storage class as default

The default storage class is responsible for all persistent volume claims which don't explicitly request `storageClassName`.

1. List the storage classes in your cluster:

    ```shell
    kubectl get storageclass
    ```

    The output is similar to this:
    ```shell
    NAME                   PROVISIONER                     AGE
    unencrypted (default)  pd.csi.storage.gke.io           1d
    encrypted-storage      gcp.csi.confidential.cloud      1d
    ```
    
    The default storage class is marked by `(default)`.

1. Mark old default storage class as non default

    If you previously used another storage class as the default, you will have to remove that annotation:
    ```shell
    kubectl patch storageclass <name-of-old-default> -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
    ```

2. Mark new class as the default

    ```shell
    kubectl patch storageclass encrypted-storage -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
    ```

3. Verify that your chosen storage class is default:

    ```shell
    kubectl get storageclass
    ```

    The output is similar to this:
    ```shell
    NAME                         PROVISIONER                     AGE
    unencrypted                  pd.csi.storage.gke.io           1d
    encrypted-storage (default)  gcp.csi.confidential.cloud      1d
    ```
