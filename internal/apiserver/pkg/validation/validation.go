// Package validation contains request-level validation rules for the iam
// apiserver.
package validation

import (
	"github.com/google/wire"

	"github.com/onexdemo/iamv2/internal/apiserver/store"
)

// Validator validates incoming API requests.
type Validator struct {
	store store.IStore
}

// ProviderSet is the Wire provider set for the validation package.
var ProviderSet = wire.NewSet(New)

// New creates a Validator with the required dependencies.
func New(ds store.IStore) *Validator {
	return &Validator{store: ds}
}
