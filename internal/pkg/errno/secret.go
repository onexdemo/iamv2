package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

var (
	// ErrSecretNotFound indicates that the requested secret does not exist.
	ErrSecretNotFound = errorsx.New(http.StatusNotFound, "NotFound.SecretNotFound", "Secret not found.")

	// ErrSecretReachMaxCount indicates that the user has reached the maximum
	// number of access keys.
	ErrSecretReachMaxCount = errorsx.New(
		http.StatusBadRequest,
		"InvalidArgument.SecretReachMaxCount",
		"Secret count has reached the maximum allowed limit.",
	)

	// ErrSecretExpired indicates that the access key has expired.
	ErrSecretExpired = errorsx.New(http.StatusUnauthorized, "Unauthenticated.SecretExpired", "The access key has expired.")

	// ErrSecretInvalid indicates that the supplied access key is invalid.
	ErrSecretInvalid = errorsx.New(http.StatusUnauthorized, "Unauthenticated.SecretInvalid", "The access key is invalid.")
)
