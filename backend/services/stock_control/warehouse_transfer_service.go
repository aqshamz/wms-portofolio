package stockcontrol

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

func warehouseTransferID() (string, error) {
	data := make([]byte, 10)
	if _, err := rand.Read(data); err != nil { return "", err }
	return "TRF-" + strings.ToUpper(hex.EncodeToString(data)), nil
}
func mapWarehouseTransfer(row repository.WarehouseTransferRow) dto.WarehouseTransferResponse {
	return dto.WarehouseTransferResponse{ID: row.ID, StatusCode: row.StatusCode, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName,
		SourceWarehouseID: row.SourceWarehouseID, SourceWarehouseCode: row.SourceWarehouseCode, SourceWarehouseName: row.SourceWarehouseName,
		TargetWarehouseID: row.TargetWarehouseID, TargetWarehouseCode: row.TargetWarehouseCode, TargetWarehouseName: row.TargetWarehouseName,
		BusinessDate: row.BusinessDate.Format("2006-01-02"), Notes: row.Notes, Lines: []dto.WarehouseTransferLineResponse{},
		ApprovedAt: row.ApprovedAt, DispatchedAt: row.DispatchedAt, ReceivedAt: row.ReceivedAt, CancelledAt: row.CancelledAt,
		CancellationReason: row.CancellationReason, CreatedAt: row.CreatedAt, CreatedBy: row.CreatedBy, CreatedByDisplayName: row.CreatedByDisplayName, VersionNo: row.VersionNo}
}
func mapWarehouseTransferLine(row repository.WarehouseTransferLineRow) dto.WarehouseTransferLineResponse {
	return dto.WarehouseTransferLineResponse{ID: row.ID, LineNo: row.LineNo, SourceBalanceID: row.SourceBalanceID, SourceBalanceVersionNo: row.SourceBalanceVersionNo,
		ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName, LotID: row.LotID, LotNumber: row.LotNumber,
		SerialID: row.SerialID, SerialNumber: row.SerialNumber, HandlingUnitID: row.HandlingUnitID, HandlingUnitBarcode: row.HandlingUnitBarcode,
		SourceLocationID: row.SourceLocationID, SourceLocationCode: row.SourceLocationCode, SourceInventoryStatusID: row.SourceInventoryStatusID,
		SourceInventoryStatusCode: row.SourceInventoryStatusCode, UOMID: row.UOMID, UOMCode: row.UOMCode, Quantity: row.Quantity,
		DispatchMovementID: row.DispatchMovementID, ReceiptLocationID: row.ReceiptLocationID, ReceiptLocationCode: row.ReceiptLocationCode,
		ReceiptMovementID: row.ReceiptMovementID, ReceivedBalanceID: row.ReceivedBalanceID, ReceivedBalanceVersionNo: row.ReceivedBalanceVersionNo,
		PutawayTargetLocationID: row.PutawayTargetLocationID, PutawayTargetLocationCode: row.PutawayTargetLocationCode,
		PutawayMovementID: row.PutawayMovementID, PutawayResultBalanceID: row.PutawayResultBalanceID, PutawayCompletedAt: row.PutawayCompletedAt}
}
func (s *Service) transferAllowed(ctx context.Context, actor, ownerID, sourceWarehouseID, targetWarehouseID string) error {
	source, err := s.repositories.Scope.Allowed(ctx, actor, ownerID, sourceWarehouseID)
	if err != nil { return err }
	target, err := s.repositories.Scope.Allowed(ctx, actor, ownerID, targetWarehouseID)
	if err != nil { return err }
	if !source && !target { return ErrForbidden }
	return nil
}
func (s *Service) GetWarehouseTransfer(ctx context.Context, id, actor string) (dto.WarehouseTransferResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 140 { return dto.WarehouseTransferResponse{}, invalid("invalid warehouse_transfer_id") }
	row, err := s.repositories.Transfers.Get(ctx, id)
	if err != nil { return dto.WarehouseTransferResponse{}, err }
	if err = s.transferAllowed(ctx, actor, row.OwnerID, row.SourceWarehouseID, row.TargetWarehouseID); err != nil { return dto.WarehouseTransferResponse{}, err }
	lines, err := s.repositories.Transfers.Lines(ctx, id)
	if err != nil { return dto.WarehouseTransferResponse{}, err }
	response := mapWarehouseTransfer(row)
	for _, line := range lines { response.Lines = append(response.Lines, mapWarehouseTransferLine(line)) }
	return response, nil
}
func (s *Service) ListWarehouseTransfers(ctx context.Context, f repository.WarehouseTransferFilter, actor string) (dto.PageResponse[dto.WarehouseTransferResponse], error) {
	f.OwnerID = strings.ToLower(strings.TrimSpace(f.OwnerID)); f.WarehouseID = strings.ToLower(strings.TrimSpace(f.WarehouseID))
	f.Side = strings.ToUpper(strings.TrimSpace(f.Side)); f.StatusCode = strings.ToUpper(strings.TrimSpace(f.StatusCode)); f.Search = strings.TrimSpace(f.Search)
	if f.OwnerID == "" || f.WarehouseID == "" || (f.Side != "SOURCE" && f.Side != "TARGET") || f.Page < 1 || f.PageSize < 1 || f.PageSize > 100 {
		return dto.PageResponse[dto.WarehouseTransferResponse]{}, invalid("invalid warehouse transfer filters")
	}
	allowed, err := s.repositories.Scope.Allowed(ctx, actor, f.OwnerID, f.WarehouseID)
	if err != nil { return dto.PageResponse[dto.WarehouseTransferResponse]{}, err }
	if !allowed { return dto.PageResponse[dto.WarehouseTransferResponse]{}, ErrForbidden }
	rows, total, err := s.repositories.Transfers.List(ctx, f)
	if err != nil { return dto.PageResponse[dto.WarehouseTransferResponse]{}, err }
	items := make([]dto.WarehouseTransferResponse, 0, len(rows))
	for _, row := range rows { items = append(items, mapWarehouseTransfer(row)) }
	return replenishmentPage(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreateWarehouseTransfer(ctx context.Context, q dto.CreateWarehouseTransferRequest, actor string) (dto.WarehouseTransferResponse, error) {
	quantity, quantityNumber, err := number(q.Quantity, true)
	if err != nil { return dto.WarehouseTransferResponse{}, err }
	date, err := time.Parse("2006-01-02", q.BusinessDate)
	if err != nil { return dto.WarehouseTransferResponse{}, invalid("business_date must be YYYY-MM-DD") }
	id, err := warehouseTransferID(); if err != nil { return dto.WarehouseTransferResponse{}, err }
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		source, err := r.Balances.Get(ctx, strings.TrimSpace(q.SourceBalanceID)); if err != nil { return err }
		if err = authorize(ctx, r, actor, source); err != nil { return err }
		if source.InventoryStatusCode != "AVAILABLE" { return invalid("warehouse transfers are limited to AVAILABLE inventory") }
		if err = validateStorageLocation(ctx, ir, source.WarehouseID, source.LocationID, "warehouse transfer source"); err != nil { return err }
		if source.HandlingUnitID != nil { return invalid("warehouse transfer of a handling unit is not supported") }
		available, ok := new(big.Rat).SetString(source.AvailableQty); if !ok || quantityNumber.Cmp(available) > 0 { return invalid("quantity exceeds available stock") }
		if source.SerialControlled {
			if quantityNumber.Cmp(big.NewRat(1, 1)) != 0 || len(q.SerialIDs) != 1 { return invalid("serialized transfers require exactly one serial and one base unit") }
		} else if len(q.SerialIDs) != 0 { return invalid("serial_ids are only allowed for serial-controlled items") }
		target, err := ir.Warehouse.GetShared(ctx, q.TargetWarehouseID); if err != nil { return err }
		if !target.IsActive { return invalid("target warehouse is inactive") }
		if source.WarehouseID == target.ID { return invalid("source and target warehouses must differ") }
		serves, err := r.Scope.WarehouseServesOwner(ctx, source.OwnerID, target.ID); if err != nil { return err }
		if !serves { return invalid("target warehouse does not actively serve the inventory owner") }
		doc, err := r.Transfers.DocumentType(ctx); if err != nil { return invalid("TRANSFER document type is not configured") }
		status, err := r.Transfers.Status(ctx, doc.ID, "DRAFT"); if err != nil { return invalid("TRANSFER DRAFT status is not configured") }
		var serialID *string
		if len(q.SerialIDs) == 1 { value := strings.TrimSpace(q.SerialIDs[0]); serialID = &value }
		header := &model.WarehouseTransfer{ID: id, DocumentTypeID: doc.ID, StatusID: status.ID, OwnerID: source.OwnerID, SourceWarehouseID: source.WarehouseID,
			TargetWarehouseID: target.ID, BusinessDate: date, Notes: q.Notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		line := &model.WarehouseTransferLine{ID: id + "-L0001", WarehouseTransferID: id, LineNo: 1, SourceBalanceID: source.ID,
			SourceBalanceVersionNo: source.VersionNo, ItemID: source.ItemID, LotID: source.LotID, SerialID: serialID, HandlingUnitID: source.HandlingUnitID,
			SourceLocationID: source.LocationID, SourceInventoryStatusID: source.InventoryStatusID, UOMID: source.UOMID, Quantity: quantity,
			CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		return r.Transfers.Create(ctx, header, line)
	})
	if err != nil { return dto.WarehouseTransferResponse{}, err }
	return s.GetWarehouseTransfer(ctx, id, actor)
}
func (s *Service) transitionWarehouseTransfer(ctx context.Context, id, expectedStatus, nextStatus, actor string, expectedVersion int64, values map[string]any) error {
	return s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		header, err := r.Transfers.Lock(ctx, id); if err != nil { return err }
		row, err := r.Transfers.Get(ctx, id); if err != nil { return err }
		if header.VersionNo != expectedVersion { return inventoryrepo.ErrConflict }
		if row.StatusCode != expectedStatus { return invalid("warehouse transfer must be " + expectedStatus) }
		allowed, err := r.Scope.Allowed(ctx, actor, header.OwnerID, header.SourceWarehouseID); if err != nil { return err }; if !allowed { return ErrForbidden }
		status, err := r.Transfers.Status(ctx, header.DocumentTypeID, nextStatus); if err != nil { return invalid("TRANSFER " + nextStatus + " status is not configured") }
		values["status_id"] = status.ID; values["updated_by"] = actor
		return r.Transfers.UpdateHeader(ctx, id, header.VersionNo, values)
	})
}
func (s *Service) ApproveWarehouseTransfer(ctx context.Context, id string, q dto.WarehouseTransferTransitionRequest, actor string) (dto.WarehouseTransferResponse, error) {
	now := time.Now(); err := s.transitionWarehouseTransfer(ctx, id, "DRAFT", "APPROVED", actor, q.ExpectedVersion, map[string]any{"approved_at": now, "approved_by": actor})
	if err != nil { return dto.WarehouseTransferResponse{}, err }; return s.GetWarehouseTransfer(ctx, id, actor)
}
func (s *Service) DispatchWarehouseTransfer(ctx context.Context, id string, q dto.WarehouseTransferTransitionRequest, actor string) (dto.WarehouseTransferResponse, error) {
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.Transfers.Lock(ctx, id); if err != nil { return err }
		row, err := r.Transfers.Get(ctx, id); if err != nil { return err }
		if header.VersionNo != q.ExpectedVersion { return inventoryrepo.ErrConflict }
		if row.StatusCode != "APPROVED" { return invalid("warehouse transfer must be APPROVED") }
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.SourceWarehouseID)); err != nil { return err }
		line, err := r.Transfers.LockLine(ctx, id); if err != nil { return err }
		serials := []string{}; if line.SerialID != nil { serials = append(serials, *line.SerialID) }
		posting := inventorydto.PostingRequest{OperationKey: "stock.transfer." + id + ".dispatch", MovementTypeCode: "TRANSFER_OUT", OwnerID: header.OwnerID,
			WarehouseID: header.SourceWarehouseID, BusinessDate: header.BusinessDate.Format("2006-01-02"), ItemID: line.ItemID, LotID: line.LotID,
			Quantity: line.Quantity, SerialIDs: serials, From: &inventorydto.BalanceDimension{LocationID: line.SourceLocationID, InventoryStatusID: line.SourceInventoryStatusID},
			ExpectedSourceVersion: &line.SourceBalanceVersionNo, SourceDocumentID: id, SourceLineID: &line.ID, Notes: header.Notes}
		posted, err := inventory.NewService(ir).PostMovement(ctx, posting, actor); if err != nil { return err }
		if err = r.Transfers.UpdateLine(ctx, line.ID, line.VersionNo, map[string]any{"dispatch_movement_id": posted.Movement.ID, "updated_by": actor}); err != nil { return err }
		status, err := r.Transfers.Status(ctx, header.DocumentTypeID, "IN_TRANSIT"); if err != nil { return invalid("TRANSFER IN_TRANSIT status is not configured") }
		now := time.Now(); return r.Transfers.UpdateHeader(ctx, id, header.VersionNo, map[string]any{"status_id": status.ID, "dispatched_at": now, "dispatched_by": actor, "updated_by": actor})
	})
	if err != nil { return dto.WarehouseTransferResponse{}, err }; return s.GetWarehouseTransfer(ctx, id, actor)
}
func validateReceivingLocation(ctx context.Context, r *inventoryrepo.Repositories, warehouseID, locationID string) error {
	location, err := r.Location.GetShared(ctx, locationID); if err != nil { return err }
	if location.WarehouseID != warehouseID || !location.IsActive || location.IsLocked { return invalid("receipt location must be active, unlocked, and in the target warehouse") }
	zone, err := r.Zone.GetShared(ctx, location.ZoneID); if err != nil { return err }
	typeRow, err := r.LocationType.GetShared(ctx, location.LocationTypeID); if err != nil { return err }
	if !zone.IsActive || !typeRow.IsActive || !typeRow.AllowsReceiving { return invalid("receipt location must allow receiving") }
	return nil
}
func (s *Service) ReceiveWarehouseTransfer(ctx context.Context, id string, q dto.ReceiveWarehouseTransferRequest, actor string) (dto.WarehouseTransferResponse, error) {
	if _, err := time.Parse("2006-01-02", q.BusinessDate); err != nil { return dto.WarehouseTransferResponse{}, invalid("business_date must be YYYY-MM-DD") }
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.Transfers.Lock(ctx, id); if err != nil { return err }; row, err := r.Transfers.Get(ctx, id); if err != nil { return err }
		if header.VersionNo != q.ExpectedVersion { return inventoryrepo.ErrConflict }; if row.StatusCode != "IN_TRANSIT" { return invalid("warehouse transfer must be IN_TRANSIT") }
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.TargetWarehouseID)); err != nil { return err }
		if err = validateReceivingLocation(ctx, ir, header.TargetWarehouseID, q.ReceiptLocationID); err != nil { return err }
		if err = validateStorageLocation(ctx, ir, header.TargetWarehouseID, q.PutawayTargetLocationID, "transfer putaway target"); err != nil { return err }
		line, err := r.Transfers.LockLine(ctx, id); if err != nil { return err }; pending, err := r.Transfers.InventoryStatus(ctx, "PUTAWAY_PENDING"); if err != nil { return invalid("PUTAWAY_PENDING inventory status is not configured") }
		serials := []string{}; if line.SerialID != nil { serials = append(serials, *line.SerialID) }
		posting := inventorydto.PostingRequest{OperationKey: "stock.transfer." + id + ".receive", MovementTypeCode: "TRANSFER_IN", OwnerID: header.OwnerID,
			WarehouseID: header.TargetWarehouseID, BusinessDate: q.BusinessDate, ItemID: line.ItemID, LotID: line.LotID, Quantity: line.Quantity, SerialIDs: serials,
			To: &inventorydto.BalanceDimension{LocationID: q.ReceiptLocationID, InventoryStatusID: pending.ID}, SourceDocumentID: id, SourceLineID: &line.ID, Notes: header.Notes}
		posted, err := inventory.NewService(ir).PostMovement(ctx, posting, actor); if err != nil { return err }
		if posted.ToBalance == nil { return fmt.Errorf("transfer receipt did not return a destination balance") }
		if err = r.Transfers.UpdateLine(ctx, line.ID, line.VersionNo, map[string]any{"receipt_location_id": q.ReceiptLocationID, "receipt_movement_id": posted.Movement.ID,
			"received_balance_id": posted.ToBalance.ID, "putaway_target_location_id": q.PutawayTargetLocationID, "updated_by": actor}); err != nil { return err }
		status, err := r.Transfers.Status(ctx, header.DocumentTypeID, "RECEIVED"); if err != nil { return invalid("TRANSFER RECEIVED status is not configured") }
		now := time.Now(); return r.Transfers.UpdateHeader(ctx, id, header.VersionNo, map[string]any{"status_id": status.ID, "received_at": now, "received_by": actor, "updated_by": actor})
	})
	if err != nil { return dto.WarehouseTransferResponse{}, err }; return s.GetWarehouseTransfer(ctx, id, actor)
}
func (s *Service) PutawayWarehouseTransfer(ctx context.Context, id string, q dto.PutawayWarehouseTransferRequest, actor string) (dto.WarehouseTransferResponse, error) {
	if _, err := time.Parse("2006-01-02", q.BusinessDate); err != nil { return dto.WarehouseTransferResponse{}, invalid("business_date must be YYYY-MM-DD") }
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.Transfers.Lock(ctx, id); if err != nil { return err }; row, err := r.Transfers.Get(ctx, id); if err != nil { return err }
		if header.VersionNo != q.ExpectedVersion { return inventoryrepo.ErrConflict }; if row.StatusCode != "RECEIVED" { return invalid("warehouse transfer must be RECEIVED") }
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.TargetWarehouseID)); err != nil { return err }
		line, err := r.Transfers.LockLine(ctx, id); if err != nil { return err }
		if line.PutawayCompletedAt != nil { return invalid("warehouse transfer is already put away") }
		if line.ReceivedBalanceID == nil || line.ReceiptLocationID == nil || line.PutawayTargetLocationID == nil { return invalid("warehouse transfer has no completed receipt") }
		balance, err := r.Balances.Get(ctx, *line.ReceivedBalanceID); if err != nil { return err }
		if balance.VersionNo != q.ExpectedBalanceVersion || balance.OwnerID != header.OwnerID || balance.WarehouseID != header.TargetWarehouseID || balance.ItemID != line.ItemID || balance.LocationID != *line.ReceiptLocationID || balance.InventoryStatusCode != "PUTAWAY_PENDING" {
			return invalid("received balance changed; refresh before putaway")
		}
		if err = validateStorageLocation(ctx, ir, header.TargetWarehouseID, *line.PutawayTargetLocationID, "transfer putaway target"); err != nil { return err }
		available, err := r.Transfers.InventoryStatus(ctx, "AVAILABLE"); if err != nil { return invalid("AVAILABLE inventory status is not configured") }
		serials := []string{}; if line.SerialID != nil { serials = append(serials, *line.SerialID) }
		posting := inventorydto.PostingRequest{OperationKey: "stock.transfer." + id + ".putaway", MovementTypeCode: "PUTAWAY", OwnerID: header.OwnerID,
			WarehouseID: header.TargetWarehouseID, BusinessDate: q.BusinessDate, ItemID: line.ItemID, LotID: line.LotID, Quantity: line.Quantity, SerialIDs: serials,
			From: &inventorydto.BalanceDimension{LocationID: *line.ReceiptLocationID, InventoryStatusID: balance.InventoryStatusID},
			To: &inventorydto.BalanceDimension{LocationID: *line.PutawayTargetLocationID, InventoryStatusID: available.ID}, ExpectedSourceVersion: &q.ExpectedBalanceVersion,
			SourceDocumentID: id, SourceLineID: &line.ID, Notes: header.Notes}
		posted, err := inventory.NewService(ir).PostMovement(ctx, posting, actor); if err != nil { return err }; if posted.ToBalance == nil { return fmt.Errorf("transfer putaway did not return a destination balance") }
		now := time.Now(); if err = r.Transfers.UpdateLine(ctx, line.ID, line.VersionNo, map[string]any{"putaway_movement_id": posted.Movement.ID, "putaway_result_balance_id": posted.ToBalance.ID,
			"putaway_completed_at": now, "putaway_completed_by": actor, "updated_by": actor}); err != nil { return err }
		return r.Transfers.UpdateHeader(ctx, id, header.VersionNo, map[string]any{"updated_by": actor})
	})
	if err != nil { return dto.WarehouseTransferResponse{}, err }; return s.GetWarehouseTransfer(ctx, id, actor)
}
func (s *Service) CancelWarehouseTransfer(ctx context.Context, id string, q dto.CancelWarehouseTransferRequest, actor string) (dto.WarehouseTransferResponse, error) {
	reason := strings.TrimSpace(q.Reason); if reason == "" { return dto.WarehouseTransferResponse{}, invalid("reason is required") }
	row, err := s.repositories.Transfers.Get(ctx, id); if err != nil { return dto.WarehouseTransferResponse{}, err }
	if row.StatusCode != "DRAFT" && row.StatusCode != "APPROVED" { return dto.WarehouseTransferResponse{}, invalid("only DRAFT or APPROVED transfers can be cancelled") }
	now := time.Now(); err = s.transitionWarehouseTransfer(ctx, id, row.StatusCode, "CANCELLED", actor, q.ExpectedVersion, map[string]any{"cancelled_at": now, "cancelled_by": actor, "cancellation_reason": reason})
	if err != nil { return dto.WarehouseTransferResponse{}, err }; return s.GetWarehouseTransfer(ctx, id, actor)
}
