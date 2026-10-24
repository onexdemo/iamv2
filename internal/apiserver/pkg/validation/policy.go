package validation

import (
	"context"

	"github.com/ory/ladon"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// ValidateCreatePolicyRequest validates a policy creation request.
func (v *Validator) ValidateCreatePolicyRequest(_ context.Context, rq *v1.CreatePolicyRequest) error {
	return validatePolicy(rq.Name, &rq.Policy)
}

// ValidateUpdatePolicyRequest validates a policy update request.
func (v *Validator) ValidateUpdatePolicyRequest(_ context.Context, rq *v1.UpdatePolicyRequest) error {
	return validatePolicy(rq.Name, &rq.Policy)
}

// ValidateGetPolicyRequest validates a single-policy lookup.
func (v *Validator) ValidateGetPolicyRequest(_ context.Context, rq *v1.GetPolicyRequest) error {
	return validateResourceName(rq.Name)
}

// ValidateDeletePolicyRequest validates a single-policy deletion.
func (v *Validator) ValidateDeletePolicyRequest(_ context.Context, rq *v1.DeletePolicyRequest) error {
	return validateResourceName(rq.Name)
}

// ValidateDeletePoliciesRequest validates a batch deletion.
func (v *Validator) ValidateDeletePoliciesRequest(_ context.Context, rq *v1.DeletePoliciesRequest) error {
	if len(rq.Names) == 0 {
		return errno.ErrInvalidArgument.WithMessage("at least one policy name is required")
	}

	for _, name := range rq.Names {
		if err := validateResourceName(name); err != nil {
			return err
		}
	}

	return nil
}

// ValidateListPoliciesRequest validates list query parameters.
func (v *Validator) ValidateListPoliciesRequest(_ context.Context, rq *v1.ListPoliciesRequest) error {
	return validateListOptions(rq.ListOptions)
}

func validatePolicy(name string, policy *ladon.DefaultPolicy) error {
	if err := validateResourceName(name); err != nil {
		return err
	}

	if policy.Effect != ladon.AllowAccess && policy.Effect != ladon.DenyAccess {
		return errno.ErrPolicyInvalid.WithMessage("effect must be %q or %q", ladon.AllowAccess, ladon.DenyAccess)
	}

	if len(policy.Subjects) == 0 || len(policy.Resources) == 0 || len(policy.Actions) == 0 {
		return errno.ErrPolicyInvalid.WithMessage("subjects, resources and actions must not be empty")
	}

	return nil
}
