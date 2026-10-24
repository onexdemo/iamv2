// Package authorization evaluates ladon policies for inbound authorization
// requests.
package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ory/ladon"

	"github.com/onexdemo/iamv2/internal/authz/pkg/analytics"
	authzv1 "github.com/onexdemo/iamv2/pkg/api/authz/v1"
)

// PolicyGetter returns the policies owned by a user.
type PolicyGetter interface {
	GetPolicy(key string) ([]*ladon.DefaultPolicy, error)
}

// Interface is the policy source and audit sink used by the ladon warden.
type Interface interface {
	List(username string) ([]*ladon.DefaultPolicy, error)
	LogRejectedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies)
	LogGrantedAccessRequest(request *ladon.Request, pool ladon.Policies, deciders ladon.Policies)
}

// Authorization implements Interface on top of an in-memory policy cache.
type Authorization struct {
	getter   PolicyGetter
	recorder *analytics.Recorder
	detailed bool
}

// NewAuthorization creates an Authorization instance.
func NewAuthorization(getter PolicyGetter, recorder *analytics.Recorder, detailed bool) *Authorization {
	return &Authorization{getter: getter, recorder: recorder, detailed: detailed}
}

// List returns every policy owned by username.
func (a *Authorization) List(username string) ([]*ladon.DefaultPolicy, error) {
	return a.getter.GetPolicy(username)
}

// LogRejectedAccessRequest records a denied authorization decision.
func (a *Authorization) LogRejectedAccessRequest(r *ladon.Request, p ladon.Policies, d ladon.Policies) {
	var conclusion string

	if len(d) > 1 {
		allowed := joinPoliciesNames(d[0 : len(d)-1])
		denied := d[len(d)-1].GetID()
		conclusion = fmt.Sprintf("policies %s allow access, but policy %s forcefully denied it", allowed, denied)
	} else if len(d) == 1 {
		conclusion = fmt.Sprintf("policy %s forcefully denied the access", d[len(d)-1].GetID())
	} else {
		conclusion = "no policy allowed access"
	}

	a.record(r, p, d, ladon.DenyAccess, conclusion)
}

// LogGrantedAccessRequest records an allowed authorization decision.
func (a *Authorization) LogGrantedAccessRequest(r *ladon.Request, p ladon.Policies, d ladon.Policies) {
	conclusion := fmt.Sprintf("policies %s allow access", joinPoliciesNames(d))
	a.record(r, p, d, ladon.AllowAccess, conclusion)
}

func (a *Authorization) record(r *ladon.Request, p, d ladon.Policies, effect, conclusion string) {
	username, _ := r.Context["username"].(string)

	requestJSON, _ := json.Marshal(r)
	policiesJSON, _ := json.Marshal(p)
	decidersJSON, _ := json.Marshal(d)

	record := &analytics.Record{
		TimeStamp:  time.Now().Unix(),
		Username:   username,
		Effect:     effect,
		Conclusion: conclusion,
		Request:    string(requestJSON),
		Policies:   string(policiesJSON),
		Deciders:   string(decidersJSON),
	}
	record.SetExpiry(0)

	if !a.detailed {
		record.Policies = ""
		record.Deciders = ""
	}

	if a.recorder != nil {
		if err := a.recorder.Record(context.Background(), record); err != nil {
			slog.Warn("failed to enqueue analytics record", "error", err)
		}
	}
}

func joinPoliciesNames(policies ladon.Policies) string {
	names := make([]string, 0, len(policies))
	for _, policy := range policies {
		names = append(names, policy.GetID())
	}

	return strings.Join(names, ", ")
}

// Authorizer evaluates ladon policies.
type Authorizer struct {
	warden ladon.Warden
}

// NewAuthorizer builds a ladon warden backed by the given policy client.
func NewAuthorizer(client Interface) *Authorizer {
	return &Authorizer{
		warden: &ladon.Ladon{
			Manager:     NewPolicyManager(client),
			AuditLogger: NewAuditLogger(client),
		},
	}
}

// Authorize evaluates a subject access review.
func (a *Authorizer) Authorize(request *ladon.Request) *authzv1.AuthzResponse {
	slog.Debug("evaluating authorization request", "subject", request.Subject, "resource", request.Resource, "action", request.Action)

	if err := a.warden.IsAllowed(request); err != nil {
		return &authzv1.AuthzResponse{Denied: true, Reason: err.Error()}
	}

	return &authzv1.AuthzResponse{Allowed: true}
}
