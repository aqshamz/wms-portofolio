package authentication

import "time"

type AuthenticationPolicy struct {
	ID                string    `gorm:"column:authentication_policy_id;type:uuid;default:gen_random_uuid();primaryKey" json:"authentication_policy_id"`
	Code              string    `gorm:"column:code;size:40;not null;uniqueIndex" json:"code"`
	Name              string    `gorm:"column:name;size:100;not null" json:"name"`
	MaxFailedAttempts int       `gorm:"column:max_failed_attempts;not null" json:"max_failed_attempts"`
	LockoutSeconds    int       `gorm:"column:lockout_seconds;not null" json:"lockout_seconds"`
	SessionTTLSeconds int       `gorm:"column:session_ttl_seconds;not null" json:"session_ttl_seconds"`
	IsDefault         bool      `gorm:"column:is_default;not null;default:false" json:"is_default"`
	IsActive          bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
}

func (AuthenticationPolicy) TableName() string { return "authentication_policy" }
