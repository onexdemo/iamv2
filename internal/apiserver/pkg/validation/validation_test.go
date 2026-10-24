package validation_test

import (
	"context"
	"testing"

	"github.com/ory/ladon"
	"github.com/stretchr/testify/require"

	"github.com/onexdemo/iamv2/internal/apiserver/pkg/validation"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

func TestValidateCreateUserRequest(t *testing.T) {
	validator := validation.New(nil)
	ctx := context.Background()

	require.NoError(t, validator.ValidateCreateUserRequest(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	}))

	require.Error(t, validator.ValidateCreateUserRequest(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "a"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	}))

	require.Error(t, validator.ValidateCreateUserRequest(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "short",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	}))

	require.Error(t, validator.ValidateCreateUserRequest(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "not-an-email",
	}))
}

func TestValidatePolicy(t *testing.T) {
	validator := validation.New(nil)
	ctx := context.Background()

	require.NoError(t, validator.ValidateCreatePolicyRequest(ctx, &v1.CreatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read"},
		Policy: ladon.DefaultPolicy{
			Subjects:  []string{"users:alice"},
			Resources: []string{"resources:articles:<.*>"},
			Actions:   []string{"get"},
			Effect:    ladon.AllowAccess,
		},
	}))

	require.Error(t, validator.ValidateCreatePolicyRequest(ctx, &v1.CreatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read"},
		Policy: ladon.DefaultPolicy{
			Subjects:  []string{"users:alice"},
			Resources: []string{"resources:articles:<.*>"},
			Actions:   []string{"get"},
			Effect:    "maybe",
		},
	}))
}
