package cryptmapper

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/edgelesssys/constellation-kms-client/pkg/config"
	"github.com/martinjungblut/go-cryptsetup"
	"k8s.io/klog"
	mount "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	cryptPrefix = "/dev/mapper/"
)

// packageLock is needed to block concurrent use of package functions, since libcryptsetup is not thread safe.
// See: https://gitlab.com/cryptsetup/cryptsetup/-/issues/710
// 		https://stackoverflow.com/questions/30553386/cryptsetup-backend-safe-with-multithreading
var packageLock = sync.Mutex{}

// CryptMapper manages dm-crypt volumes.
type CryptMapper struct {
	mapper deviceMapper
	kms    kmsClient
	kekID  string
}

// New initializes a new CryptMapper with the given kms client and key-encryption-key ID.
// kms is used to fetch data encryption keys for the dm-crypt volumes.
// kekID is the ID of the key used to encrypt the data encryption keys.
func New(kms kmsClient, kekID string, mapper deviceMapper) *CryptMapper {
	return &CryptMapper{
		mapper: mapper,
		kms:    kms,
		kekID:  kekID,
	}
}

// deviceMapper is an interface for device mapper methods.
type deviceMapper interface {
	// Init initializes a crypt device backed by 'devicePath'.
	// Sets the devieMapper to the newly allocated Device or returns any error encountered.
	// C equivalent: crypt_init
	Init(devicePath string) error
	// ActivateByVolumeKey activates a device by using a volume key.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_activate_by_volume_key
	ActivateByVolumeKey(deviceName, volumeKey string, volumeKeySize, flags int) error
	// Deactivate deactivates a device.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_deactivate
	Deactivate(deviceName string) error
	// Format formats a Device, using a specific device type, and type-independent parameters.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_format
	Format(deviceType cryptsetup.DeviceType, genericParams cryptsetup.GenericParams) error
	// Free releases crypt device context and used memory.
	// C equivalent: crypt_free
	Free() bool
	// Load loads crypt device parameters from the on-disk header.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_load
	Load() error
}

// cryptDevice is a wrapper for cryptsetup.Device.
type CryptDevice struct {
	*cryptsetup.Device
}

// Init initializes a crypt device backed by 'devicePath'.
// Sets the cryptDevice's deviceMapper to the newly allocated Device or returns any error encountered.
// C equivalent: crypt_init
func (m *CryptDevice) Init(devicePath string) error {
	device, err := cryptsetup.Init(devicePath)
	if err != nil {
		return err
	}
	m.Device = device
	return nil
}

// Free releases crypt device context and used memory.
// C equivalent: crypt_free
func (m *CryptDevice) Free() bool {
	res := m.Device.Free()
	m.Device = nil
	return res
}

// CloseCryptDevice closes the crypt device mapped for volumeID.
// Returns nil if the volume does not exist.
func (c *CryptMapper) CloseCryptDevice(volumeID string) error {
	source, err := filepath.EvalSymlinks(cryptPrefix + volumeID)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			klog.V(4).Infof("Skipping unmapping for disk %q: volume does not exist or is already unmapped", volumeID)
			return nil
		}
		return fmt.Errorf("failed to get device path for disk %q: %w", cryptPrefix+volumeID, err)
	}
	return closeCryptDevice(source, volumeID, c.mapper)
}

// closeCryptDevice closes the crypt device mapped for volumeID.
func closeCryptDevice(source, volumeID string, device deviceMapper) error {
	packageLock.Lock()
	defer packageLock.Unlock()

	klog.V(4).Infof("Unmapping dm-crypt volume %q for device %q", cryptPrefix+volumeID, source)

	if err := device.Init(source); err != nil {
		klog.Errorf("Could not initialize dm-crypt to unmap device %q: %s", source, err)
		return fmt.Errorf("could not initialize dm-crypt to unmap device %q: %w", source, err)
	}
	defer device.Free()

	if err := device.Deactivate(volumeID); err != nil {
		klog.Errorf("Could not deactivate dm-crypt volume %q for device %q: %s", cryptPrefix+volumeID, source, err)
		return fmt.Errorf("could not deactivate dm-crypt volume %q for device %q: %w", cryptPrefix+volumeID, source, err)
	}

	klog.V(4).Infof("Successfully unmapped dm-crypt volume %q for device %q", cryptPrefix+volumeID, source)
	return nil
}

// OpenCryptDevice maps the volume at source to the crypt device identified by volumeID.
// The key used to encrypt the volume is fetched using CryptMapper's kms client.
func (c *CryptMapper) OpenCryptDevice(ctx context.Context, source, volumeID string) (string, error) {
	klog.V(4).Infof("Fetching data encryption key for volume %q", volumeID)
	dek, err := c.kms.GetDEK(ctx, c.kekID, volumeID)
	if err != nil {
		return "", err
	}

	m := &mount.SafeFormatAndMount{Exec: utilexec.New()}
	return openCryptDevice(source, volumeID, string(dek), c.mapper, m.GetDiskFormat)
}

// openCryptDevice maps the volume at source to the crypt device identified by volumeID.
func openCryptDevice(source, volumeID, dek string, device deviceMapper, diskInfo func(disk string) (string, error)) (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	if len(dek) != config.SymmetricKeyLength {
		return "", fmt.Errorf("invalid key length: expected [%d], got [%d]", config.SymmetricKeyLength, len(dek))
	}

	klog.V(4).Infof("Mapping device %q to dm-crypt volume %q", source, cryptPrefix+volumeID)
	cryptsetup.SetLogCallback(func(level int, message string) { klog.V(4).Infof("libcryptsetup: %s", message) })

	// Initialize the block device
	if err := device.Init(source); err != nil {
		klog.Errorf("Initializing dm-crypt to map device %q: %s", source, err)
		return "", fmt.Errorf("initializing dm-crypt to map device %q: %w", source, err)
	}
	defer device.Free()

	// Try to load LUKS headers
	// If this fails, the device is either not formatted at all, or already formatted with a different FS
	if err := device.Load(); err != nil {
		klog.V(4).Infof("Device %q is not formatted as LUKS2 partition, checking for existing format...", source)
		format, err := diskInfo(source)
		if err != nil {
			return "", fmt.Errorf("could not determine if disk is formatted: %w", err)
		}
		if format != "" {
			// Device is already formated, return an error
			klog.Errorf("Disk %q is already formatted as: %s", source, format)
			return "", fmt.Errorf("disk %q is already formatted as: %s", source, format)
		}

		// Device is not formatted, so we can savily create a new LUKS2 partition
		klog.V(4).Infof("Device %q is not formatted. Creating new LUKS2 partition...", source)
		if err := device.Format(
			cryptsetup.LUKS2{
				SectorSize: 4096,
				PBKDFType: &cryptsetup.PbkdfType{
					Type:   "pbkdf2",
					Hash:   "sha256",
					TimeMs: 2000,
				},
			},
			cryptsetup.GenericParams{
				Cipher:        "aes",
				CipherMode:    "cbc-essiv:sha256",
				VolumeKey:     dek,
				VolumeKeySize: config.SymmetricKeyLength,
			}); err != nil {
			klog.Errorf("Formatting device %q failed: %s", source, err)
			return "", fmt.Errorf("formatting device %q failed: %w", source, err)
		}
	}

	klog.V(4).Infof("Activating LUKS2 device %q", cryptPrefix+volumeID)

	if err := device.ActivateByVolumeKey(volumeID, dek, config.SymmetricKeyLength, 0); err != nil {
		klog.Errorf("Trying to activate dm-crypt volume: %s", err)
		return "", fmt.Errorf("trying to activate dm-crypt volume: %w", err)
	}

	klog.V(4).Infof("Device %q successfully mapped to dm-crypt volume %q", source, cryptPrefix+volumeID)

	return cryptPrefix + volumeID, nil
}
