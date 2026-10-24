// Package v1 contains the public REST API types for the iam apiserver.
//
// The types in this package are the contract between iam-apiserver, iamctl and
// the authz server. Storage models live in internal/apiserver/model and are
// converted to these types before crossing the HTTP boundary.
package v1

import (
	"time"
)

// ObjectMeta is the common metadata embedded in every iam resource.
type ObjectMeta struct {
	// ID is the internal auto-increment identifier.
	ID int64 `json:"id,omitempty"`
	// InstanceID is the stable, externally visible identifier of the resource.
	InstanceID string `json:"instanceID,omitempty"`
	// Name is the user-visible resource name.
	Name string `json:"name,omitempty"`
	// Extend carries arbitrary user-defined key-value metadata.
	Extend map[string]string `json:"extend,omitempty"`
	// CreatedAt is the resource creation timestamp.
	CreatedAt time.Time `json:"createdAt,omitempty"`
	// UpdatedAt is the resource last-modification timestamp.
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// ListMeta is returned by every list endpoint.
type ListMeta struct {
	TotalCount int64 `json:"totalCount"`
}

// ListOptions is the common query parameters accepted by list endpoints.
type ListOptions struct {
	// Offset is the number of records to skip. A negative value means "no limit"
	// when combined with Limit, matching the semantics of the original iam API.
	Offset int64 `json:"offset,omitempty" form:"offset"`
	// Limit is the maximum number of records to return.
	Limit int64 `json:"limit,omitempty" form:"limit"`
	// Name filters records by a name substring.
	Name string `json:"name,omitempty" form:"name"`
}
