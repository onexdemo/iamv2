// Package v1 contains the public REST API types for the iam authz server.
package v1

// AuthzResponse is the result of a subject access review.
type AuthzResponse struct {
	// Allowed is true when the request is permitted.
	Allowed bool `json:"allowed"`
	// Denied is true when a policy explicitly denied the request.
	Denied bool `json:"denied"`
	// Reason explains why the request was allowed or denied.
	Reason string `json:"reason,omitempty"`
	// Error reports an internal error raised while evaluating the request.
	Error string `json:"error,omitempty"`
}
