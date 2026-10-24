package v1

// Secret is the public representation of an access key pair.
//
// The SecretKey is only returned when the secret is created or when it is
// explicitly fetched by its owner. List responses omit it.
type Secret struct {
	ObjectMeta `json:",inline"`

	// Username is the owner of the secret.
	Username string `json:"username"`
	// SecretID is the access key identifier.
	SecretID string `json:"secretID"`
	// SecretKey is the access key secret.
	SecretKey string `json:"secretKey,omitempty"`
	// Expires is the UNIX timestamp after which the key is no longer valid.
	// A value of 0 means the key never expires.
	Expires int64 `json:"expires"`
	// Description describes the intended use of the key.
	Description string `json:"description"`
}

// CreateSecretRequest is the payload accepted by POST /v1/secrets.
type CreateSecretRequest struct {
	ObjectMeta `json:",inline"`

	Expires     int64  `json:"expires"`
	Description string `json:"description"`
}

// UpdateSecretRequest is the payload accepted by PUT /v1/secrets/:name.
type UpdateSecretRequest struct {
	ObjectMeta `json:",inline"`

	Expires     int64  `json:"expires"`
	Description string `json:"description"`
}

// GetSecretRequest identifies a single secret.
type GetSecretRequest struct {
	Name string `json:"name" binding:"required" uri:"name"`
}

// DeleteSecretRequest identifies a single secret to delete.
type DeleteSecretRequest struct {
	Name string `json:"name" binding:"required" uri:"name"`
}

// DeleteSecretsRequest deletes several secrets in one request.
type DeleteSecretsRequest struct {
	Names []string `json:"names" binding:"required" form:"name"`
}

// ListSecretsRequest is the query payload accepted by GET /v1/secrets.
type ListSecretsRequest struct {
	ListOptions
}

// ListSecretsResponse is returned by GET /v1/secrets.
type ListSecretsResponse struct {
	ListMeta `json:",inline"`

	Items []*Secret `json:"items"`
}

// SecretInfo is the wire representation used by the authz server cache.
type SecretInfo struct {
	SecretID    string `json:"secretID"`
	Username    string `json:"username"`
	SecretKey   string `json:"secretKey"`
	Expires     int64  `json:"expires"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// ListSecretInfoRequest is the query payload accepted by the internal cache API.
type ListSecretInfoRequest struct {
	Offset int64 `json:"offset,omitempty" form:"offset"`
	Limit  int64 `json:"limit,omitempty" form:"limit"`
}

// ListSecretInfoResponse is returned by the internal cache API.
type ListSecretInfoResponse struct {
	TotalCount int64         `json:"totalCount"`
	Items      []*SecretInfo `json:"items"`
}

// CreateSecretResponse is returned by POST /v1/secrets.
type CreateSecretResponse struct {
	Secret *Secret `json:"secret"`
}

// UpdateSecretResponse is returned by PUT /v1/secrets/:name.
type UpdateSecretResponse struct {
	Secret *Secret `json:"secret"`
}

// GetSecretResponse is returned by GET /v1/secrets/:name.
type GetSecretResponse struct {
	Secret *Secret `json:"secret"`
}

// DeleteSecretResponse is returned by DELETE /v1/secrets/:name.
type DeleteSecretResponse struct{}

// DeleteSecretsResponse is returned by DELETE /v1/secrets.
type DeleteSecretsResponse struct{}
