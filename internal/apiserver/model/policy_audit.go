package model

import (
	"time"
)

// TableNamePolicyAudit is the physical table name for PolicyAuditM.
const TableNamePolicyAudit = "iamv2_apiserver_policy_audit"

// PolicyAuditM records the deletion history of authorization policies.
//
// The table is filled by database triggers in the original iam deployment and
// is periodically pruned by the iam-watcher clean watcher.
type PolicyAuditM struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement:false"`

	PolicyID string `json:"policy_id" gorm:"column:policy_id;type:varchar(64);not null;comment:Policy ID"`
	Name     string `json:"name" gorm:"column:name;type:varchar(64);not null;comment:Policy name"`
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;comment:Policy owner"`

	PolicyShadow string `json:"policy_shadow" gorm:"column:policy_shadow;type:longtext;not null;comment:JSON encoded ladon policy"`

	Extend map[string]string `json:"extend" gorm:"column:extend;type:json;serializer:json;comment:User-defined metadata"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:Creation time"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:Last modification time"`

	// DeletedAt is the time at which the original policy was deleted.
	DeletedAt time.Time `json:"deleted_at" gorm:"column:deleted_at;not null;index:idx_policy_audit_deleted_at;comment:Deletion time"`
}

// TableName returns the physical table name for PolicyAuditM.
func (*PolicyAuditM) TableName() string {
	return TableNamePolicyAudit
}
