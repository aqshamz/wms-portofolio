package inventory

import "time"

type SerialNumber struct {
	ID        string    `gorm:"column:serial_id;size:160;primaryKey"`
	OwnerID   string    `gorm:"column:owner_id;type:uuid;not null"`
	ItemID    string    `gorm:"column:item_id;type:uuid;not null"`
	SerialNo  string    `gorm:"column:serial_no;size:120;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy *string   `gorm:"column:created_by;type:uuid"`
}

func (SerialNumber) TableName() string { return "serial_number" }
