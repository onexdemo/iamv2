package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

var (
	// ErrPolicyNotFound indicates that the requested policy does not exist.
	ErrPolicyNotFound = errorsx.New(http.StatusNotFound, "NotFound.PolicyNotFound", "Policy not found.")

	// ErrPolicyAlreadyExists indicates that a policy with the same name already
	// exists for the user.
	ErrPolicyAlreadyExists = errorsx.New(http.StatusBadRequest, "AlreadyExist.PolicyAlreadyExists", "Policy already exists.")

	// ErrPolicyInvalid indicates that the ladon policy is malformed.
	ErrPolicyInvalid = errorsx.New(http.StatusBadRequest, "InvalidArgument.PolicyInvalid", "The policy is invalid.")
)
