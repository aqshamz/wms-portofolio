package authentication

type SessionRevocationReason struct {
	ID          string  `gorm:"column:session_revocation_reason_id;type:uuid;default:gen_random_uuid();primaryKey" json:"session_revocation_reason_id"`
	Code        string  `gorm:"column:code;size:40;not null;uniqueIndex" json:"code"`
	Name        string  `gorm:"column:name;size:100;not null" json:"name"`
	Description *string `gorm:"column:description;type:text" json:"description,omitempty"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true" json:"is_active"`
}

func (SessionRevocationReason) TableName() string { return "session_revocation_reason" }
