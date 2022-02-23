package cryptmapper

import (
	"context"

	"github.com/edgelesssys/constellation-kms-client/pkg/kms"
)

type kmsClient interface {
	GetDEK(ctx context.Context, kekID, dekID string, dekSize int) ([]byte, error)
	CreateKEK(ctx context.Context, keyID string, kek []byte, policyProducer kms.KeyPolicyProducer) ([]byte, error)
}
