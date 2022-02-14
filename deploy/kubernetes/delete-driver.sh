#!/bin/bash

# This script will remove the Google Compute Engine Persistent Disk CSI Driver
# from the currently available Kubernetes cluster
#
# Args:
# GCE_PD_DRIVER_VERSION: The kustomize overlay to deploy (located under
#   deploy/kubernetes/overlays).

set -o nounset
set -o errexit

readonly NAMESPACE="${GCE_PD_DRIVER_NAMESPACE:-gce-pd-csi-driver}"
readonly DEPLOY_VERSION="${GCE_PD_DRIVER_VERSION:-stable-master}"
readonly PKGDIR="$(dirname "$(readlink -f "$0")")/../.."
source "${PKGDIR}/deploy/common.sh"

ensure_kustomize

kustomize build "${PKGDIR}/deploy/kubernetes/overlays/${DEPLOY_VERSION}" | ${KUBECTL} delete -v="${VERBOSITY}" --ignore-not-found -f -
${KUBECTL} delete secret cloud-sa -v="${VERBOSITY}" --ignore-not-found

if [[ "${NAMESPACE}" != "" && "${NAMESPACE}" != "default" ]] && \
  ${KUBECTL} get namespace "${NAMESPACE}" -v="${VERBOSITY}";
then
    ${KUBECTL} delete namespace "${NAMESPACE}" -v="${VERBOSITY}"
fi
