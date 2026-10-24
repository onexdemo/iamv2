package authorization

import (
	"log/slog"

	"github.com/ory/ladon"
)

// AuditLogger writes authorization decisions to the audit sink.
type AuditLogger struct {
	client Interface
}

// NewAuditLogger creates an AuditLogger.
func NewAuditLogger(client Interface) *AuditLogger {
	return &AuditLogger{client: client}
}

// LogRejectedAccessRequest records a denied decision.
func (a *AuditLogger) LogRejectedAccessRequest(r *ladon.Request, p ladon.Policies, d ladon.Policies) {
	a.client.LogRejectedAccessRequest(r, p, d)
	slog.Debug("subject access review rejected", "subject", r.Subject, "resource", r.Resource, "action", r.Action)
}

// LogGrantedAccessRequest records an allowed decision.
func (a *AuditLogger) LogGrantedAccessRequest(r *ladon.Request, p ladon.Policies, d ladon.Policies) {
	a.client.LogGrantedAccessRequest(r, p, d)
	slog.Debug("subject access review granted", "subject", r.Subject, "resource", r.Resource, "action", r.Action)
}
