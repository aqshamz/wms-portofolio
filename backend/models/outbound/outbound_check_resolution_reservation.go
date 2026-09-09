package outbound

type OutboundCheckResolutionReservation struct {
	OutboundCheckResolutionID string `gorm:"column:outbound_check_resolution_id;size:190;primaryKey"`
	ReservationID             string `gorm:"column:reservation_id;size:140;primaryKey;uniqueIndex"`
}

func (OutboundCheckResolutionReservation) TableName() string {
	return "outbound_check_resolution_reservation"
}
