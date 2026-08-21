package authentication

import "time"

type AppSession struct {
	ID                        string     `gorm:"column:session_id;size:120;primaryKey" json:"session_id"`
	AccountID                 string     `gorm:"column:account_id;type:uuid;not null;index" json:"account_id"`
	TokenHash                 string     `gorm:"column:token_hash;size:128;not null;uniqueIndex" json:"-"`
	IssuedAt                  time.Time  `gorm:"column:issued_at;not null;default:clock_timestamp()" json:"issued_at"`
	ExpiresAt                 time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	LastSeenAt                time.Time  `gorm:"column:last_seen_at;not null;default:clock_timestamp()" json:"last_seen_at"`
	RevokedAt                 *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	SessionRevocationReasonID *string    `gorm:"column:session_revocation_reason_id;type:uuid;index" json:"session_revocation_reason_id,omitempty"`
	IPAddress                 *string    `gorm:"column:ip_address;type:inet" json:"ip_address,omitempty"`
	UserAgent                 *string    `gorm:"column:user_agent;type:text" json:"user_agent,omitempty"`
}

func (AppSession) TableName() string { return "app_session" }
