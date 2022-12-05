# v1.8.0 - Changelog since v.1.7.3

## Changes by Kind

### Feature

- Add support for setting snapshot labels ([#1017](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/1017), [@sagor999](https://github.com/sagor999))
- Go builder updated from 1.18.4 to 1.19.1. ([#1048](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/1048), [@kon-angelo](https://github.com/kon-angelo))

### Bug or Regression

- Disable devices in node unstage prior to detaching. ([#1051](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/1051), [@mattcary](https://github.com/mattcary))
- Enforce implicit pagination limit of 500 of the ListVolumesResponse#Entry field when ListVolumesRequest#max_entries is not set ([#999](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/999), [@pwschuurman](https://github.com/pwschuurman))
- Fixed issue where Regional disks are repeatedly queued for re-attaching and consuming api quota ([#1050](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/1050), [@leiyiz](https://github.com/leiyiz))

### Uncategorized

- Migrate from github.com/golang/protobuf to google.golang.org/protobuf ([#1027](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/pull/1027), [@leiyiz](https://github.com/leiyiz))
