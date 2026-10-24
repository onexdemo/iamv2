package v1

import "github.com/ory/ladon"

// Policy is the public representation of a ladon authorization policy.
type Policy struct {
	ObjectMeta `json:",inline"`

	// Username is the owner of the policy.
	Username string `json:"username"`
	// Policy is the structured ladon policy.
	Policy ladon.DefaultPolicy `json:"policy"`
	// PolicyShadow is the JSON representation stored in the database.
	// It is populated by the server and is read-only for clients.
	PolicyShadow string `json:"-"`

	// GrantTime is the time at which the policy was granted.
	GrantTime string `json:"grantTime,omitempty"`
}

// CreatePolicyRequest is the payload accepted by POST /v1/policies.
type CreatePolicyRequest struct {
	ObjectMeta `json:",inline"`

	Policy ladon.DefaultPolicy `json:"policy" binding:"required"`
}

// UpdatePolicyRequest is the payload accepted by PUT /v1/policies/:name.
type UpdatePolicyRequest struct {
	ObjectMeta `json:",inline"`

	Policy ladon.DefaultPolicy `json:"policy" binding:"required"`
}

// GetPolicyRequest identifies a single policy.
type GetPolicyRequest struct {
	Name string `json:"name" binding:"required" uri:"name"`
}

// DeletePolicyRequest identifies a single policy to delete.
type DeletePolicyRequest struct {
	Name string `json:"name" binding:"required" uri:"name"`
}

// DeletePoliciesRequest deletes several policies in one request.
type DeletePoliciesRequest struct {
	Names []string `json:"names" binding:"required" form:"name"`
}

// ListPoliciesRequest is the query payload accepted by GET /v1/policies.
type ListPoliciesRequest struct {
	ListOptions
}

// ListPoliciesResponse is returned by GET /v1/policies.
type ListPoliciesResponse struct {
	ListMeta `json:",inline"`

	Items []*Policy `json:"items"`
}

// PolicyInfo is the wire representation used by the authz server cache.
type PolicyInfo struct {
	Name         string `json:"name"`
	Username     string `json:"username"`
	PolicyShadow string `json:"policyShadow"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

// ListPolicyInfoRequest is the query payload accepted by the internal cache API.
type ListPolicyInfoRequest struct {
	Offset int64 `json:"offset,omitempty" form:"offset"`
	Limit  int64 `json:"limit,omitempty" form:"limit"`
}

// ListPolicyInfoResponse is returned by the internal cache API.
type ListPolicyInfoResponse struct {
	TotalCount int64         `json:"totalCount"`
	Items      []*PolicyInfo `json:"items"`
}

// CreatePolicyResponse is returned by POST /v1/policies.
type CreatePolicyResponse struct {
	Policy *Policy `json:"policy"`
}

// UpdatePolicyResponse is returned by PUT /v1/policies/:name.
type UpdatePolicyResponse struct {
	Policy *Policy `json:"policy"`
}

// GetPolicyResponse is returned by GET /v1/policies/:name.
type GetPolicyResponse struct {
	Policy *Policy `json:"policy"`
}

// DeletePolicyResponse is returned by DELETE /v1/policies/:name.
type DeletePolicyResponse struct{}

// DeletePoliciesResponse is returned by DELETE /v1/policies.
type DeletePoliciesResponse struct{}
