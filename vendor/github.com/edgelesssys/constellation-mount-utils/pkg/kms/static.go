package kms

import (
	"context"
	"errors"

	"github.com/edgelesssys/constellation-kms-client/pkg/kms"
)

// staticKMS is a KMS only returning keys containing of 0x41 bytes for every request.
// Use for testing ONLY.
type staticKMS struct{}

// NewStaticKMS creates a new StaticKMS.
// Use for testing ONLY.
func NewStaticKMS() *staticKMS {
	return &staticKMS{}
}

// GetDEK returns the key of staticKMS.
func (k *staticKMS) GetDEK(ctx context.Context, kekID, dekID string, dekSize int) ([]byte, error) {
	key := make([]byte, dekSize)
	for i := range key {
		key[i] = 0x41
	}
	return key, nil
}

// CreateKEK implements the kmsClient interface.
// Not implemented for StaticKMS.
func (k *staticKMS) CreateKEK(ctx context.Context, keyID string, kek []byte, policyProducer kms.KeyPolicyProducer) ([]byte, error) {
	return nil, errors.New("not implemented for StaticKMS")
}
