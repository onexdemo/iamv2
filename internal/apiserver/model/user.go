package model

import (
	"time"
)

// TableNameUser is the physical table name for UserM.
const TableNameUser = "iamv2_apiserver_user"

// UserM is the persistence model for an iam user.
type UserM struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement:true"`

	// UserID is the stable, externally visible identifier of the user.
	UserID string `json:"user_id" gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:idx_user_user_id;comment:Unique user ID"`

	// Username is the unique login name of the user.
	Username string `json:"username" gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx_user_username;comment:Username (unique)"`

	// Password is the bcrypt hash of the user's password.
	Password string `json:"-" gorm:"column:password;type:varchar(255);not null;comment:User password (hashed)"`

	// Nickname is the display name of the user.
	Nickname string `json:"nickname" gorm:"column:nickname;type:varchar(64);not null;default:'';comment:User nickname"`

	// Email is the user's email address.
	Email string `json:"email" gorm:"column:email;type:varchar(128);not null;default:'';comment:User email address"`

	// Phone is the user's phone number.
	Phone string `json:"phone" gorm:"column:phone;type:varchar(32);not null;default:'';comment:User phone number"`

	// Status is 1 for enabled and 0 for disabled accounts.
	Status int `json:"status" gorm:"column:status;not null;default:1;comment:Account status (1 enabled, 0 disabled)"`

	// IsAdmin reports whether the user has administrative privileges.
	IsAdmin int `json:"is_admin" gorm:"column:is_admin;not null;default:0;comment:Whether the user is an administrator"`

	// LoginedAt is the last successful login time. It is nil until the first login.
	LoginedAt *time.Time `json:"logined_at" gorm:"column:logined_at;default:null;comment:Last login time"`

	// Extend carries arbitrary user-defined metadata.
	Extend map[string]string `json:"extend" gorm:"column:extend;type:json;serializer:json;comment:User-defined metadata"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:Creation time"`

	// UpdatedAt is the last-modification timestamp.
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:Last modification time"`
}

// TableName returns the physical table name for UserM.
func (*UserM) TableName() string {
	return TableNameUser
}
