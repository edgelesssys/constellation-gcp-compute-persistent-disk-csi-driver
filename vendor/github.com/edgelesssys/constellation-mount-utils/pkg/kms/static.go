package kms

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation-kms-client/pkg/kms"
)

// StaticKMS is a KMS only returning one key for every request.
// Use for testing ONLY.
type StaticKMS struct {
	masterKey [32]byte
}

// NewStaticKMS creates a new StaticKMS.
// Use for testing ONLY.
func NewStaticKMS(masterKey [32]byte) *StaticKMS {
	return &StaticKMS{
		masterKey: masterKey,
	}
}

// GetDEK returns the key of StaticKMS.
func (k *StaticKMS) GetDEK(ctx context.Context, kekID, dekID string) ([]byte, error) {
	return k.masterKey[:], nil
}

// CreateKEK implements the kmsClient interface.
// Not implemented for StaticKMS.
func (k *StaticKMS) CreateKEK(ctx context.Context, keyID string, kek []byte, policyProducer kms.KeyPolicyProducer) ([]byte, error) {
	return nil, errors.New("not implemented for StaticKMS")
}
