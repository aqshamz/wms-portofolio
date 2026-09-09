package outbound

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type ShipmentDriverRow struct {
	model.ShipmentDriver
	DriverCode, DriverName, PhoneNumber, LicenseNumber string
}
type ShipmentDriverRepository struct{ db *gorm.DB }

func NewShipmentDriverRepository(db *gorm.DB) *ShipmentDriverRepository {
	return &ShipmentDriverRepository{db: db}
}
func (r *ShipmentDriverRepository) Assign(ctx context.Context, value *model.ShipmentDriver) error {
	if value.IsPrimary {
		if err := r.db.WithContext(ctx).Model(&model.ShipmentDriver{}).
			Where("shipment_id=?", value.ShipmentID).Update("is_primary", false).Error; err != nil {
			return Error(err)
		}
	}
	return Error(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "shipment_id"}, {Name: "driver_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_primary", "assigned_at", "assigned_by"}),
	}).Create(value).Error)
}
func (r *ShipmentDriverRepository) List(ctx context.Context, shipmentID string) ([]ShipmentDriverRow, error) {
	var values []ShipmentDriverRow
	err := r.db.WithContext(ctx).Table("shipment_driver a").
		Select("a.*,d.code driver_code,d.name driver_name,COALESCE(d.phone_number,'') phone_number,COALESCE(d.license_number,'') license_number").
		Joins("JOIN carrier_driver d ON d.driver_id=a.driver_id").
		Where("a.shipment_id=?", shipmentID).Order("a.is_primary DESC,d.name").Find(&values).Error
	return values, Error(err)
}
func (r *ShipmentDriverRepository) Remove(ctx context.Context, shipmentID, driverID string) error {
	result := r.db.WithContext(ctx).Where("shipment_id=? AND driver_id=?", shipmentID, driverID).
		Delete(&model.ShipmentDriver{})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *ShipmentDriverRepository) HasPrimary(ctx context.Context, shipmentID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ShipmentDriver{}).Where("shipment_id=? AND is_primary", shipmentID).Count(&count).Error
	return count > 0, Error(err)
}
