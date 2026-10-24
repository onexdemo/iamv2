package model

import (
	"time"
)

// TableNameSecret is the physical table name for SecretM.
const TableNameSecret = "iamv2_apiserver_secret"

// SecretM is the persistence model for an access key pair.
type SecretM struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement:true"`

	// SecretID is the stable, externally visible identifier of the secret.
	SecretID string `json:"secret_id" gorm:"column:secret_id;type:varchar(64);not null;uniqueIndex:idx_secret_secret_id;comment:Unique secret ID"`

	// Name is the user-visible name of the secret.
	Name string `json:"name" gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_secret_username_name;comment:Secret name"`

	// Username is the owner of the secret.
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx_secret_username_name;comment:Secret owner"`

	// SecretKey is the plaintext access key secret. It is only returned on
	// creation and must never be written to logs.
	SecretKey string `json:"-" gorm:"column:secret_key;type:varchar(255);not null;comment:Access key secret"`

	// Expires is the UNIX timestamp after which the key is invalid.
	Expires int64 `json:"expires" gorm:"column:expires;type:bigint;not null;default:0;comment:Expiration UNIX timestamp (0 = never)"`

	// Description describes the intended use of the key.
	Description string `json:"description" gorm:"column:description;type:varchar(255);not null;default:'';comment:Secret description"`

	// Extend carries arbitrary user-defined metadata.
	Extend map[string]string `json:"extend" gorm:"column:extend;type:json;serializer:json;comment:User-defined metadata"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:Creation time"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:Last modification time"`
}

// TableName returns the physical table name for SecretM.
func (*SecretM) TableName() string {
	return TableNameSecret
}
