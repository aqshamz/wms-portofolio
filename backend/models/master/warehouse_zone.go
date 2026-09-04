package master

import "time"

type WarehouseZone struct {
	ID          string    `gorm:"column:zone_id;type:uuid;default:gen_random_uuid();primaryKey;uniqueIndex:uq_zone_warehouse,priority:1" json:"zone_id"`
	WarehouseID string    `gorm:"column:warehouse_id;type:uuid;not null;uniqueIndex:uq_warehouse_zone_code,priority:1;uniqueIndex:uq_zone_warehouse,priority:2" json:"warehouse_id"`
	Code        string    `gorm:"column:code;size:40;not null;uniqueIndex:uq_warehouse_zone_code,priority:2" json:"code"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	CreatedBy   *string   `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
}

func (WarehouseZone) TableName() string { return "warehouse_zone" }
