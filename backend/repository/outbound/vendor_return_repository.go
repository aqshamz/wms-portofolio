package outbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type VendorReturnRow struct {
	model.VendorReturnTransaction
	StatusCode, OwnerCode, OwnerName, VendorCode, VendorName string
	WarehouseCode, WarehouseName, ItemCode, ItemName         string
	LotNumber, SerialNo, HandlingUnitBarcode                 string
	SourceLocationCode, SourceInventoryStatusCode, UOMCode   string
	ReturnDockLocationCode, ReturnPendingStatusCode          string
	CreatedByDisplayName                                     string
	CompletedByDisplayName, CancelledByDisplayName           *string
}

type VendorReturnRepository struct{ db *gorm.DB }

func NewVendorReturnRepository(db *gorm.DB) *VendorReturnRepository {
	return &VendorReturnRepository{db: db}
}

func vendorReturnQuery(db *gorm.DB) *gorm.DB {
	return db.Table("vendor_return_transaction r").
		Select(`r.*,status.code status_code,owner.code owner_code,owner.name owner_name,
			vendor.code vendor_code,vendor.name vendor_name,warehouse.code warehouse_code,warehouse.name warehouse_name,
			item.code item_code,item.name item_name,COALESCE(lot.lot_number,'') lot_number,
			COALESCE(serial.serial_no,'') serial_no,COALESCE(hu.barcode,'') handling_unit_barcode,
			location.code source_location_code,inventory_status.code source_inventory_status_code,
			COALESCE(return_dock.code,'') return_dock_location_code,
			COALESCE(return_status.code,'') return_pending_status_code,uom.code uom_code,
			creator.display_name created_by_display_name,completer.display_name completed_by_display_name,
			canceller.display_name cancelled_by_display_name`).
		Joins("JOIN document_status status ON status.status_id=r.status_id").
		Joins("JOIN organization owner ON owner.organization_id=r.owner_id").
		Joins("JOIN business_partner vendor ON vendor.partner_id=r.vendor_id").
		Joins("JOIN warehouse ON warehouse.warehouse_id=r.warehouse_id").
		Joins("JOIN item ON item.item_id=r.item_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=r.lot_id").
		Joins("LEFT JOIN serial_number serial ON serial.serial_id=r.serial_id").
		Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=r.handling_unit_id").
		Joins("JOIN warehouse_location location ON location.location_id=r.source_location_id").
		Joins("JOIN inventory_status ON inventory_status.inventory_status_id=r.source_inventory_status_id").
		Joins("LEFT JOIN warehouse_location return_dock ON return_dock.location_id=r.return_dock_location_id").
		Joins("LEFT JOIN inventory_status return_status ON return_status.inventory_status_id=r.return_pending_status_id").
		Joins("JOIN uom ON uom.uom_id=r.uom_id").
		Joins("JOIN app_account creator ON creator.account_id=r.created_by").
		Joins("LEFT JOIN app_account completer ON completer.account_id=r.completed_by").
		Joins("LEFT JOIN app_account canceller ON canceller.account_id=r.cancelled_by")
}

func (r *VendorReturnRepository) Get(ctx context.Context, id string) (VendorReturnRow, error) {
	var value VendorReturnRow
	err := vendorReturnQuery(r.db.WithContext(ctx)).Where("r.vendor_return_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *VendorReturnRepository) List(ctx context.Context, filter ListFilter) ([]VendorReturnRow, int64, error) {
	query := vendorReturnQuery(r.db.WithContext(ctx)).Where("r.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("r.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("r.vendor_return_id ILIKE ? OR r.quarantine_case_id ILIKE ? OR item.code ILIKE ? OR vendor.code ILIKE ? OR vendor.name ILIKE ? OR COALESCE(lot.lot_number,'') ILIKE ?", like, like, like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]VendorReturnRow, 0)
	err := query.Order("r.business_date DESC,r.vendor_return_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}

func (r *VendorReturnRepository) Lock(ctx context.Context, id string) (model.VendorReturnTransaction, error) {
	var value model.VendorReturnTransaction
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("vendor_return_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *VendorReturnRepository) Complete(ctx context.Context, id, statusID, movementID, actor string, completedAt time.Time, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.VendorReturnTransaction{}).
		Where("vendor_return_id=? AND version_no=? AND completed_at IS NULL AND cancelled_at IS NULL", id, expectedVersion).
		Updates(map[string]interface{}{
			"status_id": statusID, "inventory_movement_id": movementID,
			"completed_at": completedAt, "completed_by": actor,
			"updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor,
			"version_no": gorm.Expr("version_no+1"),
		})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

func (r *VendorReturnRepository) Cancel(ctx context.Context, id, statusID, reason, actor string, cancellationMovementID *string, expectedVersion int64) error {
	changes := map[string]interface{}{
		"status_id": statusID, "cancelled_at": time.Now(), "cancelled_by": actor,
		"cancellation_reason": reason, "updated_at": gorm.Expr("clock_timestamp()"),
		"updated_by": actor, "version_no": gorm.Expr("version_no+1"),
	}
	if cancellationMovementID != nil {
		changes["cancellation_movement_id"] = *cancellationMovementID
	}
	result := r.db.WithContext(ctx).Model(&model.VendorReturnTransaction{}).
		Where("vendor_return_id=? AND version_no=? AND completed_at IS NULL AND cancelled_at IS NULL", id, expectedVersion).
		Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
