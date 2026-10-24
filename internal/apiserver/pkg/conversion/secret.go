package conversion

import (
	"github.com/onexdemo/iamv2/internal/apiserver/model"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// SecretMToSecretV1 converts a secret model to its public representation.
func SecretMToSecretV1(secretM *model.SecretM) *v1.Secret {
	if secretM == nil {
		return nil
	}

	return &v1.Secret{
		ID:          secretM.ID,
		InstanceID:  secretM.SecretID,
		Name:        secretM.Name,
		Extend:      secretM.Extend,
		CreatedAt:   secretM.CreatedAt,
		UpdatedAt:   secretM.UpdatedAt,
		Username:    secretM.Username,
		SecretID:    secretM.SecretID,
		SecretKey:   secretM.SecretKey,
		Expires:     secretM.Expires,
		Description: secretM.Description,
	}
}

// SecretMToSecretInfo converts a secret model to the wire representation used
// by the authz server cache.
func SecretMToSecretInfo(secretM *model.SecretM) *v1.SecretInfo {
	if secretM == nil {
		return nil
	}

	return &v1.SecretInfo{
		SecretID:    secretM.SecretID,
		Username:    secretM.Username,
		SecretKey:   secretM.SecretKey,
		Expires:     secretM.Expires,
		Description: secretM.Description,
		CreatedAt:   FormatTime(secretM.CreatedAt),
		UpdatedAt:   FormatTime(secretM.UpdatedAt),
	}
}

// SecretMListToSecretV1List converts a slice of secret models to public
// representations.
func SecretMListToSecretV1List(secrets []*model.SecretM) []*v1.Secret {
	out := make([]*v1.Secret, 0, len(secrets))
	for _, secret := range secrets {
		out = append(out, SecretMToSecretV1(secret))
	}

	return out
}
