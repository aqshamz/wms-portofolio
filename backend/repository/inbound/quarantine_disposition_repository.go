package inbound

import (
	"context"
	"time"

	"gorm.io/gorm"
	model "wms-api/models/inbound"
)

type QuarantineDispositionRow struct {
	model.QuarantineDisposition
	StatusCode, DispositionTypeCode string
}

type QuarantineDispositionRepository struct{ db *gorm.DB }

func NewQuarantineDispositionRepository(db *gorm.DB) *QuarantineDispositionRepository {
	return &QuarantineDispositionRepository{db: db}
}
func (r *QuarantineDispositionRepository) Create(ctx context.Context, value *model.QuarantineDisposition) error {
	return Error(r.db.WithContext(ctx).Create(value).Error)
}
func quarantineDispositionQuery(db *gorm.DB) *gorm.DB {
	return db.Table("quarantine_disposition disposition").Select("disposition.*,status.code status_code,kind.code disposition_type_code").Joins("JOIN document_status status ON status.status_id=disposition.status_id").Joins("JOIN quarantine_disposition_type kind ON kind.quarantine_disposition_type_id=disposition.quarantine_disposition_type_id")
}
func (r *QuarantineDispositionRepository) Get(ctx context.Context, id string) (QuarantineDispositionRow, error) {
	var value QuarantineDispositionRow
	err := quarantineDispositionQuery(r.db.WithContext(ctx)).Where("disposition.quarantine_disposition_id=?", id).Take(&value).Error
	return value, Error(err)
}
func (r *QuarantineDispositionRepository) ListByCase(ctx context.Context, caseID string) ([]QuarantineDispositionRow, error) {
	rows := make([]QuarantineDispositionRow, 0)
	err := quarantineDispositionQuery(r.db.WithContext(ctx)).Where("disposition.quarantine_case_id=?", caseID).Order("disposition.decided_at,disposition.quarantine_disposition_id").Find(&rows).Error
	return rows, Error(err)
}
func (r *QuarantineDispositionRepository) Process(ctx context.Context, id, statusID, movementID string, balanceID *string) error {
	changes := map[string]interface{}{"status_id": statusID, "processed_at": time.Now(), "inventory_movement_id": movementID}
	if balanceID != nil {
		changes["resulting_balance_id"] = *balanceID
	}
	result := r.db.WithContext(ctx).Model(&model.QuarantineDisposition{}).Where("quarantine_disposition_id=? AND processed_at IS NULL", id).Updates(changes)
	if result.Error != nil {
		return Error(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrConflict
	}
	return nil
}
