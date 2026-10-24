package validation

import (
	"context"
	"strings"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// ValidateCreateSecretRequest validates a secret creation request.
func (v *Validator) ValidateCreateSecretRequest(_ context.Context, rq *v1.CreateSecretRequest) error {
	return validateSecretFields(rq.Name, rq.Expires, rq.Description)
}

// ValidateUpdateSecretRequest validates a secret update request.
func (v *Validator) ValidateUpdateSecretRequest(_ context.Context, rq *v1.UpdateSecretRequest) error {
	return validateSecretFields(rq.Name, rq.Expires, rq.Description)
}

// ValidateGetSecretRequest validates a single-secret lookup.
func (v *Validator) ValidateGetSecretRequest(_ context.Context, rq *v1.GetSecretRequest) error {
	return validateResourceName(rq.Name)
}

// ValidateDeleteSecretRequest validates a single-secret deletion.
func (v *Validator) ValidateDeleteSecretRequest(_ context.Context, rq *v1.DeleteSecretRequest) error {
	return validateResourceName(rq.Name)
}

// ValidateDeleteSecretsRequest validates a batch deletion.
func (v *Validator) ValidateDeleteSecretsRequest(_ context.Context, rq *v1.DeleteSecretsRequest) error {
	if len(rq.Names) == 0 {
		return errno.ErrInvalidArgument.WithMessage("at least one secret name is required")
	}

	for _, name := range rq.Names {
		if err := validateResourceName(name); err != nil {
			return err
		}
	}

	return nil
}

// ValidateListSecretsRequest validates list query parameters.
func (v *Validator) ValidateListSecretsRequest(_ context.Context, rq *v1.ListSecretsRequest) error {
	return validateListOptions(rq.ListOptions)
}

func validateSecretFields(name string, expires int64, description string) error {
	if err := validateResourceName(name); err != nil {
		return err
	}

	if expires < 0 {
		return errno.ErrInvalidArgument.WithMessage("expires must be greater than or equal to 0")
	}

	if len(description) > 255 {
		return errno.ErrInvalidArgument.WithMessage("description must be at most 255 characters")
	}

	return nil
}

func validateResourceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return errno.ErrInvalidArgument.WithMessage("name must be between 1 and 64 characters")
	}

	return nil
}
