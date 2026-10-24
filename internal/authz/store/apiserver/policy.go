package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ory/ladon"

	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

type policies struct {
	*datastore
}

// List returns every authorization policy grouped by owner.
func (p *policies) List(ctx context.Context) (map[string][]*ladon.DefaultPolicy, error) {
	var resp v1.ListPolicyInfoResponse

	if err := p.get(ctx, "/v1/cache/policies?limit=-1", &resp); err != nil {
		return nil, err
	}

	policies := make(map[string][]*ladon.DefaultPolicy, len(resp.Items))
	for _, item := range resp.Items {
		policy := &ladon.DefaultPolicy{}
		if err := json.Unmarshal([]byte(item.PolicyShadow), policy); err != nil {
			slog.WarnContext(ctx, "failed to decode policy, skipping", "policy", item.Name, "error", err)

			continue
		}

		policies[item.Username] = append(policies[item.Username], policy)
	}

	if resp.TotalCount > int64(len(resp.Items)) {
		return nil, fmt.Errorf("policy snapshot truncated: got %d of %d", len(resp.Items), resp.TotalCount)
	}

	return policies, nil
}
