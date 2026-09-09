package outbound

type ShipmentPacking struct {
	ShipmentID string `gorm:"column:shipment_id;size:120;primaryKey"`
	PackingID  string `gorm:"column:packing_id;size:120;primaryKey"`
}

func (ShipmentPacking) TableName() string { return "shipment_packing" }
