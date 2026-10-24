package apiserver

import (
	"context"
	"fmt"

	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

type secrets struct {
	*datastore
}

// List returns every access key indexed by secret ID.
func (s *secrets) List(ctx context.Context) (map[string]*v1.SecretInfo, error) {
	var resp v1.ListSecretInfoResponse

	if err := s.get(ctx, "/v1/cache/secrets?limit=-1", &resp); err != nil {
		return nil, err
	}

	secrets := make(map[string]*v1.SecretInfo, len(resp.Items))
	for _, item := range resp.Items {
		secrets[item.SecretID] = item
	}

	if resp.TotalCount > int64(len(secrets)) {
		return nil, fmt.Errorf("secret snapshot truncated: got %d of %d", len(secrets), resp.TotalCount)
	}

	return secrets, nil
}
