package outbound

type OutboundWaveType struct {
	ID          string  `gorm:"column:outbound_wave_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name        string  `gorm:"column:name;size:100;not null"`
	Description *string `gorm:"column:description;type:text"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
}

func (OutboundWaveType) TableName() string { return "outbound_wave_type" }
