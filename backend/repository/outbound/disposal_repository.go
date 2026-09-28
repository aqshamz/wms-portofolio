package outbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	model "wms-api/models/outbound"
)

type DisposalRow struct {
	model.DisposalTransaction
	StatusCode, OwnerCode, OwnerName, WarehouseCode, WarehouseName string
	ItemCode, ItemName, LotNumber, SerialNo                        string
	HandlingUnitBarcode, SourceLocationCode                        string
	SourceInventoryStatusCode, UOMCode, CreatedByDisplayName       string
	CompletedByDisplayName, CancelledByDisplayName                 *string
}

type DisposalRepository struct{ db *gorm.DB }

func NewDisposalRepository(db *gorm.DB) *DisposalRepository {
	return &DisposalRepository{db: db}
}

func disposalQuery(db *gorm.DB) *gorm.DB {
	return db.Table("disposal_transaction d").
		Select(`d.*,status.code status_code,owner.code owner_code,owner.name owner_name,
			warehouse.code warehouse_code,warehouse.name warehouse_name,item.code item_code,item.name item_name,
			COALESCE(lot.lot_number,'') lot_number,COALESCE(serial.serial_no,'') serial_no,
			COALESCE(hu.barcode,'') handling_unit_barcode,location.code source_location_code,
			inventory_status.code source_inventory_status_code,uom.code uom_code,
			creator.display_name created_by_display_name,completer.display_name completed_by_display_name,
			canceller.display_name cancelled_by_display_name`).
		Joins("JOIN document_status status ON status.status_id=d.status_id").
		Joins("JOIN organization owner ON owner.organization_id=d.owner_id").
		Joins("JOIN warehouse ON warehouse.warehouse_id=d.warehouse_id").
		Joins("JOIN item ON item.item_id=d.item_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=d.lot_id").
		Joins("LEFT JOIN serial_number serial ON serial.serial_id=d.serial_id").
		Joins("LEFT JOIN handling_unit hu ON hu.handling_unit_id=d.handling_unit_id").
		Joins("JOIN warehouse_location location ON location.location_id=d.source_location_id").
		Joins("JOIN inventory_status ON inventory_status.inventory_status_id=d.source_inventory_status_id").
		Joins("JOIN uom ON uom.uom_id=d.uom_id").
		Joins("JOIN app_account creator ON creator.account_id=d.created_by").
		Joins("LEFT JOIN app_account completer ON completer.account_id=d.completed_by").
		Joins("LEFT JOIN app_account canceller ON canceller.account_id=d.cancelled_by")
}

func (r *DisposalRepository) Get(ctx context.Context, id string) (DisposalRow, error) {
	var value DisposalRow
	err := disposalQuery(r.db.WithContext(ctx)).Where("d.disposal_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *DisposalRepository) List(ctx context.Context, filter ListFilter) ([]DisposalRow, int64, error) {
	query := disposalQuery(r.db.WithContext(ctx)).Where("d.owner_id=?", filter.OwnerID)
	if filter.WarehouseID != "" {
		query = query.Where("d.warehouse_id=?", filter.WarehouseID)
	}
	if filter.StatusCode != "" {
		query = query.Where("status.code=?", filter.StatusCode)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("d.disposal_id ILIKE ? OR d.quarantine_case_id ILIKE ? OR item.code ILIKE ? OR COALESCE(lot.lot_number,'') ILIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]DisposalRow, 0)
	err := query.Order("d.business_date DESC,d.disposal_id DESC").Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}

func (r *DisposalRepository) Lock(ctx context.Context, id string) (model.DisposalTransaction, error) {
	var value model.DisposalTransaction
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("disposal_id=?", id).Take(&value).Error
	return value, Error(err)
}

func (r *DisposalRepository) Complete(ctx context.Context, id, statusID, movementID, actor string, completedAt time.Time, expectedVersion int64) error {
	result := r.db.WithContext(ctx).Model(&model.DisposalTransaction{}).
		Where("disposal_id=? AND version_no=? AND completed_at IS NULL AND cancelled_at IS NULL", id, expectedVersion).
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

func (r *DisposalRepository) Cancel(ctx context.Context, id, statusID, reason, actor string, expectedVersion int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.DisposalTransaction{}).
		Where("disposal_id=? AND version_no=? AND completed_at IS NULL AND cancelled_at IS NULL", id, expectedVersion).
		Updates(map[string]interface{}{
			"status_id": statusID, "cancelled_at": now, "cancelled_by": actor,
			"cancellation_reason": reason, "updated_at": gorm.Expr("clock_timestamp()"),
			"updated_by": actor, "version_no": gorm.Expr("version_no+1"),
		})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}

type DisposalDispositionContext struct {
	DispositionDocumentTypeID string
	DispositionStatusID       string
	DispositionStatusCode     string
	CaseDocumentTypeID        string
	CaseStatusID              string
	CaseStatusCode            string
	CaseVersionNo             int64
	QuarantineQty             string
	ProcessedQty              string
	CommittedQty              string
}

func (r *DisposalRepository) LockDispositionContext(ctx context.Context, dispositionID string) (DisposalDispositionContext, error) {
	var disposition struct {
		DocumentTypeID, StatusID string
	}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Table("quarantine_disposition").Select("document_type_id,status_id").Where("quarantine_disposition_id=?", dispositionID).Take(&disposition).Error; err != nil {
		return DisposalDispositionContext{}, Error(err)
	}
	var value DisposalDispositionContext
	err := r.db.WithContext(ctx).Table("quarantine_disposition d").
		Select(`d.document_type_id disposition_document_type_id,d.status_id disposition_status_id,ds.code disposition_status_code,
			qc.document_type_id case_document_type_id,qc.status_id case_status_id,qcs.code case_status_code,qc.version_no case_version_no,
			qc.quarantine_qty::text quarantine_qty,
			COALESCE((SELECT sum(x.disposition_qty) FROM quarantine_disposition x JOIN document_status xs ON xs.status_id=x.status_id WHERE x.quarantine_case_id=qc.quarantine_case_id AND xs.code='PROCESSED'),0)::text processed_qty,
			COALESCE((SELECT sum(x.disposition_qty) FROM quarantine_disposition x JOIN document_status xs ON xs.status_id=x.status_id WHERE x.quarantine_case_id=qc.quarantine_case_id AND xs.code IN ('DECIDED','PROCESSED')),0)::text committed_qty`).
		Joins("JOIN document_status ds ON ds.status_id=d.status_id").
		Joins("JOIN quarantine_case qc ON qc.quarantine_case_id=d.quarantine_case_id").
		Joins("JOIN document_status qcs ON qcs.status_id=qc.status_id").
		Where("d.quarantine_disposition_id=?", dispositionID).Take(&value).Error
	return value, Error(err)
}

func (r *DisposalRepository) ProcessDisposition(ctx context.Context, id, statusID, movementID string) error {
	result := r.db.WithContext(ctx).Table("quarantine_disposition").Where("quarantine_disposition_id=? AND processed_at IS NULL", id).
		Updates(map[string]interface{}{"status_id": statusID, "processed_at": time.Now(), "inventory_movement_id": movementID})
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}

func (r *DisposalRepository) CancelDisposition(ctx context.Context, id, statusID string) error {
	result := r.db.WithContext(ctx).Table("quarantine_disposition").Where("quarantine_disposition_id=? AND processed_at IS NULL", id).Update("status_id", statusID)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}

func (r *DisposalRepository) UpdateCaseStatus(ctx context.Context, id, statusID, actor string, close bool, expectedVersion int64) error {
	changes := map[string]interface{}{"status_id": statusID, "updated_at": gorm.Expr("clock_timestamp()"), "updated_by": actor, "version_no": gorm.Expr("version_no+1")}
	if close {
		changes["closed_at"] = time.Now()
	} else {
		changes["closed_at"] = nil
	}
	result := r.db.WithContext(ctx).Table("quarantine_case").Where("quarantine_case_id=(SELECT quarantine_case_id FROM quarantine_disposition WHERE quarantine_disposition_id=?) AND version_no=?", id, expectedVersion).Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConcurrentWrite
	}
	return nil
}
