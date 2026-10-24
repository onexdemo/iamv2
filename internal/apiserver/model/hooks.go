package model

import (
	"github.com/onexstack/onexstack/pkg/rid"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"
)

// AfterCreate assigns the stable user ID once the database has generated the
// primary key.
func (m *UserM) AfterCreate(tx *gorm.DB) error {
	m.UserID = rid.NewResourceID("user").New(uint64(m.ID))
	return tx.Save(m).Error
}

// AfterCreate assigns the stable secret ID once the database has generated the
// primary key.
func (m *SecretM) AfterCreate(tx *gorm.DB) error {
	m.SecretID = rid.NewResourceID("secret").New(uint64(m.ID))
	return tx.Save(m).Error
}

// AfterCreate assigns the stable policy ID once the database has generated the
// primary key.
func (m *PolicyM) AfterCreate(tx *gorm.DB) error {
	m.PolicyID = rid.NewResourceID("policy").New(uint64(m.ID))
	return tx.Save(m).Error
}

func init() {
	registry.Register(&UserM{})
	registry.Register(&SecretM{})
	registry.Register(&PolicyM{})
	registry.Register(&PolicyAuditM{})
}
