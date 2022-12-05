# Copyright (c) Edgeless Systems GmbH
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, version 3 of the License.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
#
# This file incorporates work covered by the following copyright and
# permission notice:
#
#
# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Args:
# GCE_PD_CSI_STAGING_IMAGE: Staging image repository
REV=$(shell git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)
GCE_PD_CSI_STAGING_VERSION ?= ${REV}
STAGINGVERSION=${GCE_PD_CSI_STAGING_VERSION}
STAGINGIMAGE=${GCE_PD_CSI_STAGING_IMAGE}
DRIVERBINARY=gce-pd-csi-driver
DRIVERWINDOWSBINARY=${DRIVERBINARY}.exe

DOCKER=DOCKER_BUILDKIT=1 docker

BASE_IMAGE_LTSC2019=mcr.microsoft.com/windows/servercore:ltsc2019
BASE_IMAGE_20H2=mcr.microsoft.com/windows/servercore:20H2

# Both arrays MUST be index aligned.
WINDOWS_IMAGE_TAGS=ltsc2019 20H2
WINDOWS_BASE_IMAGES=$(BASE_IMAGE_LTSC2019) $(BASE_IMAGE_20H2)

GCFLAGS=""
ifdef GCE_PD_CSI_DEBUG
	GCFLAGS="all=-N -l"
endif

all: gce-pd-driver gce-pd-driver-windows
gce-pd-driver: require-GCE_PD_CSI_STAGING_VERSION
	mkdir -p bin
	go build -trimpath -gcflags=$(GCFLAGS) -ldflags "-X main.version=$(STAGINGVERSION) -s -w" -o bin/${DRIVERBINARY} ./cmd/gce-pd-csi-driver/

gce-pd-driver-windows: require-GCE_PD_CSI_STAGING_VERSION
ifeq (${GOARCH}, amd64)
	mkdir -p bin
	GOOS=windows go build -ldflags -X=main.version=$(STAGINGVERSION) -o bin/${DRIVERWINDOWSBINARY} ./cmd/gce-pd-csi-driver/
else
	$(warning gcp-pd-driver-windows only supports amd64.)
endif

build-container: require-GCE_PD_CSI_STAGING_IMAGE require-GCE_PD_CSI_STAGING_VERSION
	$(DOCKER) build --platform=linux --progress=plain \
		-t $(STAGINGIMAGE):$(STAGINGVERSION) \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .

build-and-push-windows-container-ltsc2019: require-GCE_PD_CSI_STAGING_IMAGE
	$(DOCKER) build --file=Dockerfile.Windows --platform=windows \
		-t $(STAGINGIMAGE):$(STAGINGVERSION)_ltsc2019 \
		--build-arg BASE_IMAGE=$(BASE_IMAGE_LTSC2019) \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)_ltsc2019

build-and-push-windows-container-20H2: require-GCE_PD_CSI_STAGING_IMAGE
	$(DOCKER) build --file=Dockerfile.Windows --platform=windows \
		-t $(STAGINGIMAGE):$(STAGINGVERSION)_20H2 \
		--build-arg BASE_IMAGE=$(BASE_IMAGE_20H2) \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)_20H2

build-and-push-multi-arch: build-and-push-container-linux-amd64 build-and-push-container-linux-arm64 build-and-push-windows-container-ltsc2019 build-and-push-windows-container-20H2
	$(DOCKER) manifest create --amend $(STAGINGIMAGE):$(STAGINGVERSION) $(STAGINGIMAGE):$(STAGINGVERSION)_linux_amd64 $(STAGINGIMAGE):$(STAGINGVERSION)_linux_arm64 $(STAGINGIMAGE):$(STAGINGVERSION)_20H2 $(STAGINGIMAGE):$(STAGINGVERSION)_ltsc2019
	STAGINGIMAGE="$(STAGINGIMAGE)" STAGINGVERSION="$(STAGINGVERSION)" WINDOWS_IMAGE_TAGS="$(WINDOWS_IMAGE_TAGS)" WINDOWS_BASE_IMAGES="$(WINDOWS_BASE_IMAGES)" ./manifest_osversion.sh
	$(DOCKER) manifest push -p $(STAGINGIMAGE):$(STAGINGVERSION)

build-and-push-multi-arch-debug: build-and-push-container-linux-debug build-and-push-windows-container-ltsc2019
	$(DOCKER) manifest create --amend $(STAGINGIMAGE):$(STAGINGVERSION) $(STAGINGIMAGE):$(STAGINGVERSION)_linux $(STAGINGIMAGE):$(STAGINGVERSION)_ltsc2019
	STAGINGIMAGE="$(STAGINGIMAGE)" STAGINGVERSION="$(STAGINGVERSION)" WINDOWS_IMAGE_TAGS="ltsc2019" WINDOWS_BASE_IMAGES="$(BASE_IMAGE_LTSC2019)" ./manifest_osversion.sh
	$(DOCKER) manifest push -p $(STAGINGIMAGE):$(STAGINGVERSION)

push-container: build-container
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)

# Used by hack/verify-docker-deps.sh, not used for building artifacts
validate-container-linux-amd64:
	$(DOCKER) build --platform=linux/amd64 \
		-t validation_linux_amd64 \
		--target validation-image \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .

# Used by hack/verify-docker-deps.sh, not used for building artifacts
validate-container-linux-arm64:
	$(DOCKER) build --platform=linux/arm64 \
		-t validation_linux_arm64 \
		--target validation-image \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .

build-and-push-container-linux-amd64: require-GCE_PD_CSI_STAGING_IMAGE
	$(DOCKER) build --platform=linux/amd64 \
		-t $(STAGINGIMAGE):$(STAGINGVERSION)_linux_amd64 \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)_linux_amd64

build-and-push-container-linux-arm64: require-GCE_PD_CSI_STAGING_IMAGE
	$(DOCKER) build --file=Dockerfile.arm64 --platform=linux/arm64 \
		-t $(STAGINGIMAGE):$(STAGINGVERSION)_linux_arm64 \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)_linux_arm64

build-and-push-container-linux-debug: require-GCE_PD_CSI_STAGING_IMAGE
	$(DOCKER) build --file=Dockerfile.debug --platform=linux \
		-t $(STAGINGIMAGE):$(STAGINGVERSION)_linux \
		--build-arg BUILDPLATFORM=linux \
		--build-arg STAGINGVERSION=$(STAGINGVERSION) .
	$(DOCKER) push $(STAGINGIMAGE):$(STAGINGVERSION)_linux

test-sanity: gce-pd-driver
	go test --v -timeout 30s sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/test/sanity -run ^TestSanity$

test-k8s-integration:
	go build -o bin/k8s-integration-test ./test/k8s-integration

require-GCE_PD_CSI_STAGING_IMAGE:
ifndef GCE_PD_CSI_STAGING_IMAGE
	$(error "Must set environment variable GCE_PD_CSI_STAGING_IMAGE to staging image repository")
endif

require-GCE_PD_CSI_STAGING_VERSION:
ifndef GCE_PD_CSI_STAGING_VERSION
	$(error "Must set environment variable GCE_PD_CSI_STAGING_VERSION to build a runnable driver")
endif
