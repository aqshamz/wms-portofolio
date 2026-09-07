package inventory

import (
	"context"
	"strings"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	repository "wms-api/repository/inventory"
)

func validateCoreIDs(values ...*string) error {
	for _, value := range values {
		*value = strings.TrimSpace(*value)
		if *value != "" && !uuid(*value) {
			return invalid("ID filters must be UUIDs")
		}
	}
	return nil
}
func validateIdentityFilter(value string, max int, label string) error {
	if value != "" && !identityID(value, max) {
		return invalid("invalid " + label + " filter")
	}
	return nil
}

func validateCoreText(value string, max int, label string) error {
	if len(value) > max {
		return invalid(label + " filter is too long")
	}
	return nil
}
func mapBalance(v repository.BalanceRow) dto.BalanceResponse {
	return dto.BalanceResponse{ID: v.ID, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, LocationID: v.LocationID, LocationCode: v.LocationCode, ItemID: v.ItemID, ItemCode: v.ItemCode, ItemName: v.ItemName, LotID: v.LotID, LotNumber: v.LotNumber, HandlingUnitID: v.HandlingUnitID, InventoryStatusID: v.InventoryStatusID, InventoryStatusCode: v.InventoryStatusCode, OnHandQty: v.OnHandQty, ReservedQty: v.ReservedQty, AvailableQty: v.AvailableQty, UOMID: v.UOMID, UOMCode: v.UOMCode, VersionNo: v.VersionNo, UpdatedAt: v.UpdatedAt}
}
func mapMovement(v repository.MovementRow) dto.MovementResponse {
	return dto.MovementResponse{ID: v.ID, MovementTypeID: v.MovementTypeID, MovementTypeCode: v.MovementTypeCode, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, BusinessDate: v.BusinessDate.Format("2006-01-02"), OccurredAt: v.OccurredAt, ItemID: v.ItemID, ItemCode: v.ItemCode, LotID: v.LotID, LotNumber: v.LotNumber, SerialID: v.SerialID, HandlingUnitID: v.HandlingUnitID, FromLocationID: v.FromLocationID, FromLocationCode: v.FromLocationCode, ToLocationID: v.ToLocationID, ToLocationCode: v.ToLocationCode, FromStatusID: v.FromStatusID, FromStatusCode: v.FromStatusCode, ToStatusID: v.ToStatusID, ToStatusCode: v.ToStatusCode, Quantity: v.Quantity, UOMID: v.UOMID, UOMCode: v.UOMCode, SourceDocumentID: v.SourceDocumentID, SourceLineID: v.SourceLineID, ReasonCodeID: v.ReasonCodeID, Notes: v.Notes, OperationKey: v.OperationKey, CreatedBy: v.CreatedBy}
}
func mapSerialState(v repository.SerialStateRow) dto.SerialStateResponse {
	return dto.SerialStateResponse{SerialID: v.SerialID, SerialNo: v.SerialNo, Balance: mapBalance(v.Balance), VersionNo: v.VersionNo, UpdatedAt: v.UpdatedAt}
}
func mapMovementType(v model.MovementType) dto.MovementTypeResponse {
	return dto.MovementTypeResponse{ID: v.ID, Code: v.Code, Name: v.Name, Description: v.Description, IsActive: v.IsActive}
}
func (s *Service) GetBalance(ctx context.Context, id string) (dto.BalanceResponse, error) {
	if !identityID(id, 160) {
		return dto.BalanceResponse{}, invalid("invalid balance_id")
	}
	v, err := s.repositories.Balance.Get(ctx, id)
	return mapBalance(v), err
}
func (s *Service) ListBalances(ctx context.Context, f repository.BalanceFilter) (dto.PageResponse[dto.BalanceResponse], error) {
	if err := validateCorePage(f.Page, f.PageSize); err != nil {
		return dto.PageResponse[dto.BalanceResponse]{}, err
	}
	if err := validateCoreIDs(&f.OwnerID, &f.WarehouseID, &f.LocationID, &f.ItemID, &f.InventoryStatusID); err != nil {
		return dto.PageResponse[dto.BalanceResponse]{}, err
	}
	if err := validateIdentityFilter(f.LotID, 120, "lot_id"); err != nil {
		return dto.PageResponse[dto.BalanceResponse]{}, err
	}
	if err := validateIdentityFilter(f.HandlingUnitID, 120, "handling_unit_id"); err != nil {
		return dto.PageResponse[dto.BalanceResponse]{}, err
	}
	rows, total, err := s.repositories.Balance.List(ctx, f)
	items := make([]dto.BalanceResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapBalance(v))
	}
	return corePage(items, f.Page, f.PageSize, total), err
}
func (s *Service) GetMovement(ctx context.Context, id string) (dto.MovementResponse, error) {
	if !identityID(id, 140) {
		return dto.MovementResponse{}, invalid("invalid movement_id")
	}
	v, err := s.repositories.Movement.Get(ctx, id)
	return mapMovement(v), err
}
func (s *Service) ListMovements(ctx context.Context, f repository.MovementFilter) (dto.PageResponse[dto.MovementResponse], error) {
	if err := validateCorePage(f.Page, f.PageSize); err != nil {
		return dto.PageResponse[dto.MovementResponse]{}, err
	}
	if err := validateCoreIDs(&f.OwnerID, &f.WarehouseID, &f.ItemID, &f.MovementTypeID); err != nil {
		return dto.PageResponse[dto.MovementResponse]{}, err
	}
	if f.OperationKey != "" && (len(f.OperationKey) > 160 || !operationKeyPattern.MatchString(f.OperationKey)) {
		return dto.PageResponse[dto.MovementResponse]{}, invalid("invalid operation_key filter")
	}
	if err := validateCoreText(f.SourceDocumentID, 140, "source_document_id"); err != nil {
		return dto.PageResponse[dto.MovementResponse]{}, err
	}
	rows, total, err := s.repositories.Movement.List(ctx, f)
	items := make([]dto.MovementResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapMovement(v))
	}
	return corePage(items, f.Page, f.PageSize, total), err
}
func (s *Service) GetSerialState(ctx context.Context, id string) (dto.SerialStateResponse, error) {
	if !identityID(id, 160) {
		return dto.SerialStateResponse{}, invalid("invalid serial_id")
	}
	v, err := s.repositories.SerialState.Get(ctx, id)
	return mapSerialState(v), err
}
func (s *Service) ListSerialStates(ctx context.Context, f repository.SerialStateFilter) (dto.PageResponse[dto.SerialStateResponse], error) {
	if err := validateCorePage(f.Page, f.PageSize); err != nil {
		return dto.PageResponse[dto.SerialStateResponse]{}, err
	}
	if err := validateCoreIDs(&f.OwnerID, &f.WarehouseID, &f.LocationID, &f.ItemID, &f.InventoryStatusID); err != nil {
		return dto.PageResponse[dto.SerialStateResponse]{}, err
	}
	if err := validateIdentityFilter(f.LotID, 120, "lot_id"); err != nil {
		return dto.PageResponse[dto.SerialStateResponse]{}, err
	}
	if err := validateIdentityFilter(f.HandlingUnitID, 120, "handling_unit_id"); err != nil {
		return dto.PageResponse[dto.SerialStateResponse]{}, err
	}
	rows, total, err := s.repositories.SerialState.List(ctx, f)
	items := make([]dto.SerialStateResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapSerialState(v))
	}
	return corePage(items, f.Page, f.PageSize, total), err
}
func (s *Service) ListMovementTypes(ctx context.Context, active *bool) ([]dto.MovementTypeResponse, error) {
	rows, err := s.repositories.MovementType.List(ctx, active)
	items := make([]dto.MovementTypeResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapMovementType(v))
	}
	return items, err
}
func validateCorePage(page, size int) error {
	if page < 1 || page > 1000000 || size < 1 || size > 100 {
		return invalid("invalid pagination")
	}
	return nil
}
func corePage[T any](items []T, page, size int, total int64) dto.PageResponse[T] {
	pages := (total + int64(size) - 1) / int64(size)
	return dto.PageResponse[T]{Items: items, Page: page, PageSize: size, TotalItems: total, TotalPages: pages}
}
