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

FROM ubuntu:20.04 as lib-builder
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    git gcc make autoconf automake autopoint pkg-config libtool gettext libssl-dev libdevmapper-dev \
    libpopt-dev uuid-dev libsepol1-dev libjson-c-dev libssh-dev libblkid-dev tar libargon2-0-dev libpwquality-dev

# Build libcryptsetup from source so we can disable udev support
RUN git clone -b v2.4.3 https://gitlab.com/cryptsetup/cryptsetup/ /cryptsetup
WORKDIR /cryptsetup
RUN ./autogen.sh
# Disable udev support since this causes a deadlock on Kubernetes
RUN CC=gcc CXX=g++ CFLAGS="-O1 -g" CXXFLAGS="-O1 -g" ./configure --enable-static --disable-udev
RUN make

FROM ubuntu:20.04 as builder
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y libdevmapper-dev libjson-c-dev wget pkg-config build-essential
RUN wget https://go.dev/dl/go1.17.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz
ENV PATH=${PATH}:/usr/local/go/bin

COPY --from=lib-builder /cryptsetup/.libs/libcryptsetup.so /usr/lib/x86_64-linux-gnu/libcryptsetup.so
COPY --from=lib-builder /cryptsetup/lib/libcryptsetup.pc /usr/lib/x86_64-linux-gnu/pkgconfig/libcryptsetup.pc
COPY --from=lib-builder /cryptsetup/lib/libcryptsetup.h  /usr/include/libcryptsetup.h

RUN ln -s  /usr/lib/x86_64-linux-gnu/libcryptsetup.so  /usr/lib/x86_64-linux-gnu/libcryptsetup.so.12

ARG STAGINGVERSION
ARG TARGETPLATFORM

#RUN apt-get update && apt-get install -y libcryptsetup-dev && apt-get autoremove -y && apt-get autoclean -y
WORKDIR /go/src/sigs.k8s.io/gcp-compute-persistent-disk-csi-driver
COPY . .
RUN GOARCH=$(echo $TARGETPLATFORM | cut -f2 -d '/') GCE_PD_CSI_STAGING_VERSION=$STAGINGVERSION make gce-pd-driver

# MAD HACKS: Build a version first so we can take the scsi_id bin and put it somewhere else in our real build
FROM k8s.gcr.io/build-image/debian-base:buster-v1.9.0 as mad-hack
RUN clean-install udev

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y util-linux e2fsprogs mount ca-certificates udev xfsprogs nvme-cli xxd libdevmapper-dev libjson-c-dev

COPY --from=builder /go/src/sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/bin/gce-pd-csi-driver /gce-pd-csi-driver

COPY --from=lib-builder /cryptsetup/.libs/libcryptsetup.so /usr/lib/x86_64-linux-gnu/libcryptsetup.so
RUN ln -s  /usr/lib/x86_64-linux-gnu/libcryptsetup.so  /usr/lib/x86_64-linux-gnu/libcryptsetup.so.12
COPY --from=mad-hack /lib/udev/scsi_id /lib/udev_containerized/scsi_id
COPY deploy/kubernetes/udev/google_nvme_id /lib/udev_containerized/google_nvme_id

ENTRYPOINT ["/gce-pd-csi-driver"]
