package outbound

type OutboundCheckExceptionStatus struct {
	ID       string `gorm:"column:outbound_check_exception_status_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code     string `gorm:"column:code;size:30;not null;uniqueIndex"`
	Name     string `gorm:"column:name;size:100;not null"`
	IsFinal  bool   `gorm:"column:is_final;not null;default:false"`
	IsActive bool   `gorm:"column:is_active;not null;default:true"`
}

func (OutboundCheckExceptionStatus) TableName() string { return "outbound_check_exception_status" }
