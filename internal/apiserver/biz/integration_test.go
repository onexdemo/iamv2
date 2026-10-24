package biz_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/onexstack/onexstack/pkg/db"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"github.com/onexstack/onexstack/pkg/token"
	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	policybiz "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/policy"
	secretbiz "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/secret"
	userbiz "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/user"
	"github.com/onexdemo/iamv2/internal/apiserver/model"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/notifier"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

func newStore(t *testing.T) store.IStore {
	t.Helper()

	dbInstance, err := db.NewInMemorySQLite(filepath.Join(t.TempDir(), "iam.db"))
	require.NoError(t, err)
	require.NoError(t, registry.Migrate(dbInstance))

	return store.NewStore(dbInstance)
}

func TestUserLifecycle(t *testing.T) {
	token.Init("test-signing-key", token.WithIdentityKey("userID"), token.WithExpiration(time.Hour))

	st := newStore(t)
	users := userbiz.New(st, notifier.NewNoopNotifier())
	ctx := context.Background()

	created, err := users.Create(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", created.User.Name)
	assert.NotEmpty(t, created.User.InstanceID)

	_, err = users.Create(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	})
	require.Error(t, err)

	got, err := users.Get(ctx, &v1.GetUserRequest{Username: "alice"})
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", got.User.Email)

	list, err := users.List(ctx, &v1.ListUsersRequest{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, list.TotalCount)

	login, err := users.Login(ctx, &v1.LoginRequest{Username: "alice", Password: "secret1"})
	require.NoError(t, err)
	assert.NotEmpty(t, login.Token)

	_, err = users.Login(ctx, &v1.LoginRequest{Username: "alice", Password: "wrong"})
	require.Error(t, err)
}

func TestSecretLifecycle(t *testing.T) {
	st := newStore(t)
	secrets := secretbiz.New(st, notifier.NewNoopNotifier())
	ctx := contextx.WithUsername(context.Background(), "alice")

	created, err := secrets.Create(ctx, &v1.CreateSecretRequest{
		ObjectMeta:  v1.ObjectMeta{Name: "demo"},
		Expires:     time.Now().Add(time.Hour).Unix(),
		Description: "demo key",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Secret.SecretID)
	assert.NotEmpty(t, created.Secret.SecretKey)

	got, err := secrets.Get(ctx, &v1.GetSecretRequest{Name: "demo"})
	require.NoError(t, err)
	assert.Equal(t, created.Secret.SecretKey, got.Secret.SecretKey)

	list, err := secrets.List(ctx, &v1.ListSecretsRequest{})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	assert.Empty(t, list.Items[0].SecretKey, "list responses must not expose secret material")

	_, err = secrets.Delete(ctx, &v1.DeleteSecretRequest{Name: "demo"})
	require.NoError(t, err)

	_, err = secrets.Get(ctx, &v1.GetSecretRequest{Name: "demo"})
	require.Error(t, err)
}

func TestPolicyLifecycle(t *testing.T) {
	st := newStore(t)
	policies := policybiz.New(st, notifier.NewNoopNotifier())
	ctx := contextx.WithUsername(context.Background(), "alice")

	policy := ladon.DefaultPolicy{
		Description: "read articles",
		Subjects:    []string{"users:alice"},
		Resources:   []string{"resources:articles:<.*>"},
		Actions:     []string{"get"},
		Effect:      ladon.AllowAccess,
	}

	created, err := policies.Create(ctx, &v1.CreatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read-articles"},
		Policy:     policy,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Policy.PolicyShadow)

	got, err := policies.Get(ctx, &v1.GetPolicyRequest{Name: "read-articles"})
	require.NoError(t, err)
	assert.Equal(t, "read-articles", got.Policy.Policy.ID)
	assert.Equal(t, ladon.AllowAccess, got.Policy.Policy.Effect)

	list, err := policies.List(ctx, &v1.ListPoliciesRequest{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, list.TotalCount)

	deny := policy
	deny.Effect = ladon.DenyAccess
	updated, err := policies.Update(ctx, &v1.UpdatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read-articles"},
		Policy:     deny,
	})
	require.NoError(t, err)
	assert.Equal(t, ladon.DenyAccess, updated.Policy.Policy.Effect)
}

func TestPolicyDeletionArchivesAudit(t *testing.T) {
	st := newStore(t)
	policies := policybiz.New(st, notifier.NewNoopNotifier())
	ctx := contextx.WithUsername(context.Background(), "alice")

	_, err := policies.Create(ctx, &v1.CreatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read-articles"},
		Policy: ladon.DefaultPolicy{
			Subjects:  []string{"users:alice"},
			Resources: []string{"resources:articles:<.*>"},
			Actions:   []string{"get"},
			Effect:    ladon.AllowAccess,
		},
	})
	require.NoError(t, err)

	_, err = policies.Delete(ctx, &v1.DeletePolicyRequest{Name: "read-articles"})
	require.NoError(t, err)

	var count int64
	require.NoError(t, st.DB(ctx).Model(&model.PolicyAuditM{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestUserDeletionCascades(t *testing.T) {
	st := newStore(t)
	users := userbiz.New(st, notifier.NewNoopNotifier())
	secrets := secretbiz.New(st, notifier.NewNoopNotifier())
	policies := policybiz.New(st, notifier.NewNoopNotifier())

	ctx := context.Background()
	_, err := users.Create(ctx, &v1.CreateUserRequest{
		ObjectMeta: v1.ObjectMeta{Name: "alice"},
		Password:   "secret1",
		Nickname:   "Alice",
		Email:      "alice@example.com",
	})
	require.NoError(t, err)

	userCtx := contextx.WithUsername(ctx, "alice")
	_, err = secrets.Create(userCtx, &v1.CreateSecretRequest{ObjectMeta: v1.ObjectMeta{Name: "demo"}})
	require.NoError(t, err)
	_, err = policies.Create(userCtx, &v1.CreatePolicyRequest{
		ObjectMeta: v1.ObjectMeta{Name: "read-articles"},
		Policy: ladon.DefaultPolicy{
			Subjects:  []string{"users:alice"},
			Resources: []string{"resources:articles:<.*>"},
			Actions:   []string{"get"},
			Effect:    ladon.AllowAccess,
		},
	})
	require.NoError(t, err)

	adminCtx := contextx.WithIsAdmin(ctx, true)
	_, err = users.Delete(adminCtx, &v1.DeleteUserRequest{Username: "alice"})
	require.NoError(t, err)

	_, err = users.Get(ctx, &v1.GetUserRequest{Username: "alice"})
	require.Error(t, err)

	listSecrets, err := secrets.ListAllSecrets(ctx, &v1.ListSecretInfoRequest{})
	require.NoError(t, err)
	assert.Zero(t, listSecrets.TotalCount)

	listPolicies, err := policies.ListAllPolicies(ctx, &v1.ListPolicyInfoRequest{})
	require.NoError(t, err)
	assert.Zero(t, listPolicies.TotalCount)

	var audits int64
	require.NoError(t, st.DB(ctx).Model(&model.PolicyAuditM{}).Count(&audits).Error)
	assert.EqualValues(t, 1, audits)
}
