package outbound

type DeliveryFailureReason struct {
	ID          string  `gorm:"column:delivery_failure_reason_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (DeliveryFailureReason) TableName() string { return "delivery_failure_reason" }
