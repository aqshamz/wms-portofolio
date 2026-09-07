package inventory

import (
	"context"
	dto "wms-api/dto/inventory"
	repository "wms-api/repository/inventory"
)

func (s *Service) GetLot(ctx context.Context, id string) (dto.LotResponse, error) {
	if !identityID(id, 120) {
		return dto.LotResponse{}, invalid("invalid identity ID")
	}
	v, err := s.repositories.Lot.Get(ctx, id)
	return mapLot(v), err
}
func (s *Service) ListLot(ctx context.Context, f repository.Filter) (dto.PageResponse[dto.LotResponse], error) {
	if err := validateFilter(&f, false); err != nil {
		return dto.PageResponse[dto.LotResponse]{}, err
	}
	rows, total, err := s.repositories.Lot.List(ctx, f)
	items := make([]dto.LotResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapLot(v))
	}
	return page(items, f, total), err
}
func (s *Service) GetSerial(ctx context.Context, id string) (dto.SerialResponse, error) {
	if !identityID(id, 160) {
		return dto.SerialResponse{}, invalid("invalid identity ID")
	}
	v, err := s.repositories.Serial.Get(ctx, id)
	return mapSerial(v), err
}
func (s *Service) ListSerial(ctx context.Context, f repository.Filter) (dto.PageResponse[dto.SerialResponse], error) {
	if err := validateFilter(&f, false); err != nil {
		return dto.PageResponse[dto.SerialResponse]{}, err
	}
	rows, total, err := s.repositories.Serial.List(ctx, f)
	items := make([]dto.SerialResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapSerial(v))
	}
	return page(items, f, total), err
}
func (s *Service) GetHandlingUnit(ctx context.Context, id string) (dto.HandlingUnitResponse, error) {
	if !identityID(id, 120) {
		return dto.HandlingUnitResponse{}, invalid("invalid identity ID")
	}
	v, err := s.repositories.HandlingUnit.Get(ctx, id)
	return mapHandlingUnit(v), err
}
func (s *Service) ListHandlingUnit(ctx context.Context, f repository.Filter) (dto.PageResponse[dto.HandlingUnitResponse], error) {
	if err := validateFilter(&f, true); err != nil {
		return dto.PageResponse[dto.HandlingUnitResponse]{}, err
	}
	rows, total, err := s.repositories.HandlingUnit.List(ctx, f)
	items := make([]dto.HandlingUnitResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapHandlingUnit(v))
	}
	return page(items, f, total), err
}
