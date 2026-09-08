package inbound

type QuarantineDispositionType struct {
	ID                    string  `gorm:"column:quarantine_disposition_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code                  string  `gorm:"column:code;size:40;not null"`
	Name                  string  `gorm:"column:name;size:100;not null"`
	Description           *string `gorm:"column:description;type:text"`
	ReleasesToAvailable   bool    `gorm:"column:releases_to_available;not null;default:false"`
	RequiresReinspection  bool    `gorm:"column:requires_reinspection;not null;default:false"`
	RemovesInventory      bool    `gorm:"column:removes_inventory;not null;default:false"`
	RemovalMovementTypeID *string `gorm:"column:removal_movement_type_id;type:uuid"`
	IsActive              bool    `gorm:"column:is_active;not null;default:true"`
}

func (QuarantineDispositionType) TableName() string { return "quarantine_disposition_type" }
