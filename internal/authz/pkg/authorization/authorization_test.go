package authorization_test

import (
	"testing"

	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"

	"github.com/onexdemo/iamv2/internal/authz/pkg/authorization"
)

type fakeGetter struct {
	policies []*ladon.DefaultPolicy
}

func (f fakeGetter) GetPolicy(string) ([]*ladon.DefaultPolicy, error) {
	return f.policies, nil
}

func TestAuthorizer(t *testing.T) {
	policy := &ladon.DefaultPolicy{
		ID:        "read-articles",
		Subjects:  []string{"users:alice"},
		Resources: []string{"resources:articles:<.*>"},
		Actions:   []string{"get"},
		Effect:    ladon.AllowAccess,
	}

	client := authorization.NewAuthorization(fakeGetter{policies: []*ladon.DefaultPolicy{policy}}, nil, true)
	authorizer := authorization.NewAuthorizer(client)

	allowed := authorizer.Authorize(&ladon.Request{
		Subject:  "users:alice",
		Action:   "get",
		Resource: "resources:articles:1",
		Context:  ladon.Context{"username": "alice"},
	})
	assert.True(t, allowed.Allowed)
	assert.False(t, allowed.Denied)

	denied := authorizer.Authorize(&ladon.Request{
		Subject:  "users:alice",
		Action:   "delete",
		Resource: "resources:articles:1",
		Context:  ladon.Context{"username": "alice"},
	})
	assert.False(t, denied.Allowed)
	assert.True(t, denied.Denied)
}
