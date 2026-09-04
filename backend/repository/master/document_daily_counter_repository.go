package master

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
	model "wms-api/models/master"
)

var ErrCounterExhausted = errors.New("daily document counter exhausted")

type DocumentDailyCounterRepository struct{ db *gorm.DB }

func NewDocumentDailyCounterRepository(db *gorm.DB) *DocumentDailyCounterRepository {
	return &DocumentDailyCounterRepository{db: db}
}

// Next is a single atomic upsert. At the configured limit no row is updated.
func (r *DocumentDailyCounterRepository) Next(ctx context.Context, typeID string, date time.Time, limit int64) (int64, error) {
	var value model.DocumentDailyCounter
	result := r.db.WithContext(ctx).Raw(`INSERT INTO document_daily_counter (document_type_id,business_date,last_number)
 VALUES (?, ?, 1) ON CONFLICT (document_type_id,business_date) DO UPDATE SET
 last_number=document_daily_counter.last_number+1, updated_at=clock_timestamp()
 WHERE document_daily_counter.last_number < ? RETURNING *`, typeID, date, limit).Scan(&value)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, ErrCounterExhausted
	}
	return value.LastNumber, nil
}
func (r *DocumentDailyCounterRepository) List(ctx context.Context, typeID string, from, to time.Time, page, size int) ([]model.DocumentDailyCounter, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.DocumentDailyCounter{}).Where("document_type_id = ? AND business_date BETWEEN ? AND ?", typeID, from, to)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]model.DocumentDailyCounter, 0)
	err := query.Order("business_date DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error
	return rows, total, err
}
