package store

import (
	"context"
	"time"

	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
)

// PolicyAuditStore defines the persistence operations for policy audit rows.
type PolicyAuditStore interface {
	// CreateBatch archives deleted policies.
	CreateBatch(ctx context.Context, audits []*model.PolicyAuditM) error
	// ClearOutdated removes audit rows older than maxReserveDays days.
	ClearOutdated(ctx context.Context, maxReserveDays int) (int64, error)
}

type policyAuditStore struct {
	s *store
}

var _ PolicyAuditStore = (*policyAuditStore)(nil)

func newPolicyAuditStore(s *store) *policyAuditStore {
	return &policyAuditStore{s: s}
}

// CreateBatch archives deleted policies in a single statement.
func (p *policyAuditStore) CreateBatch(ctx context.Context, audits []*model.PolicyAuditM) error {
	if len(audits) == 0 {
		return nil
	}

	return p.s.DB(ctx).WithContext(ctx).Create(&audits).Error
}

// ClearOutdated removes audit rows older than maxReserveDays days.
func (p *policyAuditStore) ClearOutdated(ctx context.Context, maxReserveDays int) (int64, error) {
	if maxReserveDays <= 0 {
		return 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -maxReserveDays)

	result := p.s.DB(ctx,
		where.NewWhere().Q("deleted_at < ?", cutoff),
	).Delete(&model.PolicyAuditM{})

	return result.RowsAffected, result.Error
}
