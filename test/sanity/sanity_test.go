/*
Copyright (c) Edgeless Systems GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, version 3 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

This file incorporates work covered by the following copyright and
permission notice:


Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sanitytest

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	sanity "github.com/kubernetes-csi/csi-test/v4/pkg/sanity"
	compute "google.golang.org/api/compute/v1"
	common "sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/common"
	"sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/deviceutils"
	gce "sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-cloud-provider/compute"
	metadataservice "sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-cloud-provider/metadata"
	driver "sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/gce-pd-csi-driver"
	mountmanager "sigs.k8s.io/gcp-compute-persistent-disk-csi-driver/pkg/mount-manager"
)

func TestSanity(t *testing.T) {
	// Set up variables
	driverName := "test-driver"
	project := "test-project"
	zone := "country-region-zone"
	vendorVersion := "test-version"
	tmpDir := "/tmp/csi"
	extraLabels := map[string]string{"test-label": "test-label-value"}
	endpoint := fmt.Sprintf("unix:%s/csi.sock", tmpDir)
	mountPath := path.Join(tmpDir, "mount")
	stagePath := path.Join(tmpDir, "stage")
	skipTests := strings.Join([]string{
		"NodeExpandVolume.*should work if node-expand is called after node-publish",
		"NodeExpandVolume.*should fail when volume is not found",
		"ListSnapshots.*should return snapshots that match the specified source volume id",
	}, "|")

	// Set up driver and env
	gceDriver := driver.GetGCEDriver()

	cloudProvider, err := gce.CreateFakeCloudProvider(project, zone, nil)
	if err != nil {
		t.Fatalf("Failed to get cloud provider: %v", err.Error())
	}

	fallbackRequisiteZones := []string{}
	enableStoragePools := false
	multiZoneVolumeHandleConfig := driver.MultiZoneVolumeHandleConfig{}
	listVolumesConfig := driver.ListVolumesConfig{}

	mounter := mountmanager.NewFakeSafeMounter()
	deviceUtils := deviceutils.NewFakeDeviceUtils(true)

	// Initialize GCE Driver
	identityServer := driver.NewIdentityServer(gceDriver)
	controllerServer := driver.NewControllerServer(gceDriver, cloudProvider, 0, 5*time.Minute, fallbackRequisiteZones, enableStoragePools, multiZoneVolumeHandleConfig, listVolumesConfig)
	fakeStatter := mountmanager.NewFakeStatterWithOptions(mounter, mountmanager.FakeStatterOptions{IsBlock: false})
	nodeServer := driver.NewNodeServer(gceDriver, mounter, deviceUtils, metadataservice.NewFakeService(), fakeStatter, &fakeCryptMapper{})
	err = gceDriver.SetupGCEDriver(driverName, vendorVersion, extraLabels, nil, identityServer, controllerServer, nodeServer)
	if err != nil {
		t.Fatalf("Failed to initialize GCE CSI Driver: %v", err.Error())
	}

	instance := &compute.Instance{
		Name:  "test-name",
		Disks: []*compute.AttachedDisk{},
	}
	cloudProvider.InsertInstance(instance, "test-location", "test-name")

	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sanity temp working dir %s: %v", tmpDir, err.Error())
	}

	defer func() {
		// Clean up tmp dir
		if err = os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to clean up sanity temp working dir %s: %v", tmpDir, err.Error())
		}
	}()

	go func() {
		gceDriver.Run(endpoint, 10000, false)
	}()

	// TODO(#818): Fix failing tests and remove test skip flag.
	flag.Set("ginkgo.skip", skipTests)

	// Run test
	config := sanity.TestConfig{
		TargetPath:     mountPath,
		StagingPath:    stagePath,
		Address:        endpoint,
		DialOptions:    []grpc.DialOption{grpc.WithInsecure()},
		IDGen:          newPDIDGenerator(project, zone),
		TestVolumeSize: common.GbToBytes(200),
	}
	sanity.Test(t, config)
}

type pdIDGenerator struct {
	project string
	zone    string
}

var _ sanity.IDGenerator = &pdIDGenerator{}

func newPDIDGenerator(project, zone string) *pdIDGenerator {
	return &pdIDGenerator{
		project: project,
		zone:    zone,
	}
}

func (p pdIDGenerator) GenerateUniqueValidVolumeID() string {
	return common.CreateZonalVolumeID(p.project, p.zone, uuid.New().String()[:10])
}

func (p pdIDGenerator) GenerateInvalidVolumeID() string {
	return "fake-volid"
}

func (p pdIDGenerator) GenerateUniqueValidNodeID() string {
	return common.CreateNodeID(p.project, p.zone, uuid.New().String()[:10])
}

func (p pdIDGenerator) GenerateInvalidNodeID() string {
	return "fake-nodeid"
}

type fakeCryptMapper struct {
	deviceName string
}

func (s *fakeCryptMapper) CloseCryptDevice(volumeID string) error {
	return nil
}

func (s *fakeCryptMapper) OpenCryptDevice(ctx context.Context, source, volumeID string, integrity bool) (string, error) {
	return "/dev/mapper/" + volumeID, nil
}

func (s *fakeCryptMapper) ResizeCryptDevice(ctx context.Context, volumeID string) (string, error) {
	return "/dev/mapper/" + volumeID, nil
}

func (s *fakeCryptMapper) GetDevicePath(volumeID string) (string, error) {
	return s.deviceName, nil
}
