package conversion

import (
	"github.com/onexdemo/iamv2/internal/apiserver/model"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// PolicyMToPolicyV1 converts a policy model to its public representation.
func PolicyMToPolicyV1(policyM *model.PolicyM) *v1.Policy {
	if policyM == nil {
		return nil
	}

	return &v1.Policy{
		ID:           policyM.ID,
		InstanceID:   policyM.PolicyID,
		Name:         policyM.Name,
		Extend:       policyM.Extend,
		CreatedAt:    policyM.CreatedAt,
		UpdatedAt:    policyM.UpdatedAt,
		Username:     policyM.Username,
		Policy:       policyM.Policy,
		PolicyShadow: policyM.PolicyShadow,
		GrantTime:    FormatTime(policyM.CreatedAt),
	}
}

// PolicyMToPolicyInfo converts a policy model to the wire representation used
// by the authz server cache.
func PolicyMToPolicyInfo(policyM *model.PolicyM) *v1.PolicyInfo {
	if policyM == nil {
		return nil
	}

	return &v1.PolicyInfo{
		Name:         policyM.Name,
		Username:     policyM.Username,
		PolicyShadow: policyM.PolicyShadow,
		CreatedAt:    FormatTime(policyM.CreatedAt),
	}
}

// PolicyMListToPolicyV1List converts a slice of policy models to public
// representations.
func PolicyMListToPolicyV1List(policies []*model.PolicyM) []*v1.Policy {
	out := make([]*v1.Policy, 0, len(policies))
	for _, policy := range policies {
		out = append(out, PolicyMToPolicyV1(policy))
	}

	return out
}
