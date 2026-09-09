package outbound

type CarrierService struct {
	ID        string `gorm:"column:carrier_service_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CarrierID string `gorm:"column:carrier_id;type:uuid;not null;uniqueIndex:uq_carrier_service_code,priority:1"`
	Code      string `gorm:"column:code;size:40;not null;uniqueIndex:uq_carrier_service_code,priority:2"`
	Name      string `gorm:"column:name;size:100;not null"`
	IsActive  bool   `gorm:"column:is_active;not null;default:true"`
}

func (CarrierService) TableName() string { return "carrier_service" }
