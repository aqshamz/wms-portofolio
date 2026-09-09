package outbound

type OutboundCheckResolutionMovement struct {
	OutboundCheckResolutionID string `gorm:"column:outbound_check_resolution_id;size:190;primaryKey"`
	MovementID                string `gorm:"column:movement_id;size:140;primaryKey;uniqueIndex"`
}

func (OutboundCheckResolutionMovement) TableName() string {
	return "outbound_check_resolution_movement"
}
