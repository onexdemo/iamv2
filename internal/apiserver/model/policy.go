package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ory/ladon"
	"gorm.io/gorm"
)

// TableNamePolicy is the physical table name for PolicyM.
const TableNamePolicy = "iamv2_apiserver_policy"

// PolicyM is the persistence model for a ladon authorization policy.
type PolicyM struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement:true"`

	// PolicyID is the stable, externally visible identifier of the policy.
	PolicyID string `json:"policy_id" gorm:"column:policy_id;type:varchar(64);not null;uniqueIndex:idx_policy_policy_id;comment:Unique policy ID"`

	// Name is the user-visible name of the policy and the ladon policy ID.
	Name string `json:"name" gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_policy_username_name;comment:Policy name"`

	// Username is the owner of the policy.
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx_policy_username_name;comment:Policy owner"`

	// PolicyShadow is the JSON encoded ladon policy. Clients must not modify it.
	PolicyShadow string `json:"-" gorm:"column:policy_shadow;type:longtext;not null;comment:JSON encoded ladon policy"`

	// Extend carries arbitrary user-defined metadata.
	Extend map[string]string `json:"extend" gorm:"column:extend;type:json;serializer:json;comment:User-defined metadata"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:Creation time"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:Last modification time"`

	// Policy is the decoded ladon policy. It is not stored in its own column.
	Policy ladon.DefaultPolicy `json:"policy" gorm:"-"`
}

// TableName returns the physical table name for PolicyM.
func (*PolicyM) TableName() string {
	return TableNamePolicy
}

// BeforeCreate encodes the ladon policy into PolicyShadow.
func (m *PolicyM) BeforeCreate(_ *gorm.DB) error {
	return m.encodePolicy()
}

// BeforeUpdate encodes the ladon policy into PolicyShadow.
func (m *PolicyM) BeforeUpdate(_ *gorm.DB) error {
	return m.encodePolicy()
}

// AfterFind decodes PolicyShadow into Policy.
func (m *PolicyM) AfterFind(_ *gorm.DB) error {
	if m.PolicyShadow == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(m.PolicyShadow), &m.Policy); err != nil {
		return fmt.Errorf("decode policy shadow: %w", err)
	}

	return nil
}

func (m *PolicyM) encodePolicy() error {
	m.Policy.ID = m.Name

	data, err := json.Marshal(m.Policy)
	if err != nil {
		return fmt.Errorf("encode policy shadow: %w", err)
	}

	m.PolicyShadow = string(data)

	return nil
}

// ToAudit converts a policy into its deletion-audit representation.
func (m *PolicyM) ToAudit() *PolicyAuditM {
	return &PolicyAuditM{
		ID:           m.ID,
		PolicyID:     m.PolicyID,
		Name:         m.Name,
		Username:     m.Username,
		PolicyShadow: m.PolicyShadow,
		Extend:       m.Extend,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    time.Now(),
	}
}
