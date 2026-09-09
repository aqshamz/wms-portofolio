package outbound

type DeliveryEventType struct {
	ID             string  `gorm:"column:delivery_event_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code           string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name           string  `gorm:"column:name;size:100;not null"`
	Description    *string `gorm:"column:description;type:text"`
	MarksDelivered bool    `gorm:"column:marks_delivered;not null;default:false"`
	MarksFailed    bool    `gorm:"column:marks_failed;not null;default:false"`
	MarksReturned  bool    `gorm:"column:marks_returned;not null;default:false"`
	IsActive       bool    `gorm:"column:is_active;not null;default:true"`
}

func (DeliveryEventType) TableName() string { return "delivery_event_type" }
