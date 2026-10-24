package authorization

import (
	"fmt"

	"github.com/ory/ladon"
)

// PolicyManager is a read-only ladon.Manager backed by the policy cache.
type PolicyManager struct {
	client Interface
}

// NewPolicyManager creates a policy manager.
func NewPolicyManager(client Interface) ladon.Manager {
	return &PolicyManager{client: client}
}

// Create is unsupported: policies are managed through iam-apiserver.
func (*PolicyManager) Create(ladon.Policy) error {
	return nil
}

// Update is unsupported: policies are managed through iam-apiserver.
func (*PolicyManager) Update(ladon.Policy) error {
	return nil
}

// Get is unsupported: policies are managed through iam-apiserver.
func (*PolicyManager) Get(string) (ladon.Policy, error) {
	return nil, nil
}

// Delete is unsupported: policies are managed through iam-apiserver.
func (*PolicyManager) Delete(string) error {
	return nil
}

// GetAll is unsupported: policies are managed through iam-apiserver.
func (*PolicyManager) GetAll(int64, int64) (ladon.Policies, error) {
	return nil, nil
}

// FindRequestCandidates returns the policies owned by the request's user.
func (m *PolicyManager) FindRequestCandidates(r *ladon.Request) (ladon.Policies, error) {
	username, _ := r.Context["username"].(string)
	if username == "" {
		return nil, fmt.Errorf("authorization request is missing the username context")
	}

	policies, err := m.client.List(username)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	out := make([]ladon.Policy, 0, len(policies))
	for _, policy := range policies {
		out = append(out, policy)
	}

	return out, nil
}

// FindPoliciesForSubject is not used by the cache-backed manager.
func (*PolicyManager) FindPoliciesForSubject(string) (ladon.Policies, error) {
	return nil, nil
}

// FindPoliciesForResource is not used by the cache-backed manager.
func (*PolicyManager) FindPoliciesForResource(string) (ladon.Policies, error) {
	return nil, nil
}
