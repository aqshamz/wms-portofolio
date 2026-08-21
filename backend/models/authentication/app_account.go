package authentication

import "time"

type AppAccount struct {
	ID                     string     `gorm:"column:account_id;type:uuid;default:gen_random_uuid();primaryKey" json:"account_id"`
	Username               string     `gorm:"column:username;size:100;not null;uniqueIndex" json:"username"`
	Email                  *string    `gorm:"column:email;size:254;uniqueIndex" json:"email,omitempty"`
	DisplayName            string     `gorm:"column:display_name;size:150;not null" json:"display_name"`
	PasswordHash           *string    `gorm:"column:password_hash;type:text" json:"-"`
	ExternalSubject        *string    `gorm:"column:external_subject;size:255;uniqueIndex" json:"-"`
	AccountStatusID        string     `gorm:"column:account_status_id;type:uuid;not null;index" json:"account_status_id"`
	AuthenticationPolicyID *string    `gorm:"column:authentication_policy_id;type:uuid;index" json:"authentication_policy_id,omitempty"`
	PreferredTimezone      *string    `gorm:"column:preferred_timezone;size:50" json:"preferred_timezone,omitempty"`
	FailedLoginCount       int        `gorm:"column:failed_login_count;not null;default:0" json:"-"`
	LockedUntil            *time.Time `gorm:"column:locked_until" json:"-"`
	LastLoginAt            *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	CreatedBy              *string    `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()" json:"updated_at"`
	UpdatedBy              *string    `gorm:"column:updated_by;type:uuid" json:"updated_by,omitempty"`
	VersionNo              int        `gorm:"column:version_no;not null;default:1" json:"version_no"`
}

func (AppAccount) TableName() string { return "app_account" }
